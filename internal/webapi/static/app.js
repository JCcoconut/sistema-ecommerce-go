const $ = (selector) => document.querySelector(selector);
const dinero = (valor) => new Intl.NumberFormat("es-EC", {style:"currency", currency:"USD"}).format(valor);

async function api(ruta, opciones = {}) {
  const respuesta = await fetch(ruta, {headers:{"Content-Type":"application/json", ...(opciones.headers || {})}, ...opciones});
  const texto = await respuesta.text();
  const datos = texto ? JSON.parse(texto) : null;
  if (!respuesta.ok) throw new Error(datos?.error?.mensaje || `Error HTTP ${respuesta.status}`);
  return datos;
}

function mensaje(texto, esError = false) {
  const caja = $("#mensaje");
  caja.textContent = texto;
  caja.classList.toggle("error", esError);
  caja.hidden = false;
  window.setTimeout(() => { caja.hidden = true; }, 3600);
}

function texto(tag, contenido, clase = "") {
  const nodo = document.createElement(tag); nodo.textContent = contenido; if (clase) nodo.className = clase; return nodo;
}

async function cargarEstado() {
  try {
    const estado = await api("/api/health");
    $("#estado-api").textContent = `${estado.productos} productos · ${estado.pedidos} pedidos · persistencia ${estado.persistencia ? "activa" : "en memoria"}`;
  } catch (error) { $("#estado-api").textContent = "Servidor no disponible"; mensaje(error.message, true); }
}

async function cargarProductos() {
  const q = encodeURIComponent($("#buscar").value.trim());
  const productos = await api(`/api/productos${q ? `?q=${q}` : ""}`);
  const contenedor = $("#productos"); contenedor.replaceChildren();
  if (!productos.length) { contenedor.append(texto("p", "No se encontraron productos.")); return; }
  productos.forEach((producto) => {
    const card = document.createElement("article"); card.className = "product";
    card.append(texto("span", producto.categoria, "tag"), texto("h3", producto.nombre));
    card.append(texto("p", `Proveedor: ${producto.proveedor.nombre}`));
    const pie = document.createElement("div"); pie.className = "product-foot";
    const datos = document.createElement("div"); datos.append(texto("div", dinero(producto.precio), "price"), texto("div", `${producto.stock} unidades disponibles`, "stock"));
    const boton = texto("button", "Agregar", "button primary"); boton.type = "button"; boton.disabled = producto.stock < 1;
    boton.addEventListener("click", async () => { try { await api("/api/carrito/items", {method:"POST", body:JSON.stringify({producto_id:producto.id,cantidad:1})}); mensaje(`${producto.nombre} agregado`); await cargarCarrito(); } catch (e) { mensaje(e.message, true); } });
    pie.append(datos, boton); card.append(pie); contenedor.append(card);
  });
}

async function cargarCarrito() {
  const carro = await api("/api/carrito");
  $("#badge-carrito").textContent = carro.cantidad_total;
  const lista = $("#items-carrito"); lista.replaceChildren();
  if (!carro.items.length) lista.append(texto("p", "Tu carrito todavía está vacío."));
  carro.items.forEach((item) => {
    const fila = document.createElement("div"); fila.className = "cart-item";
    const detalle = document.createElement("div"); detalle.append(texto("strong", item.nombre), texto("small", `${item.cantidad} × ${dinero(item.precio_unitario)}`));
    const acciones = document.createElement("div"); acciones.append(texto("strong", dinero(item.total)));
    const eliminar = texto("button", "  Eliminar", "icon-button"); eliminar.type = "button";
    eliminar.addEventListener("click", async () => { try { await api(`/api/carrito/items/${encodeURIComponent(item.producto_id)}`, {method:"DELETE"}); await cargarCarrito(); } catch (e) { mensaje(e.message, true); } });
    acciones.append(eliminar); fila.append(detalle, acciones); lista.append(fila);
  });
  const totales = $("#totales"); totales.replaceChildren();
  [["Subtotal",carro.subtotal],[`Descuento${carro.cupon ? ` (${carro.cupon})` : ""}`,-carro.descuento],["Total",carro.total]].forEach(([etiqueta,valor], indice) => {
    const fila = document.createElement("div"); if (indice === 2) fila.className = "grand"; fila.append(texto("span", etiqueta), texto("span", dinero(valor))); totales.append(fila);
  });
}

async function cargarPedidos() {
  const pedidos = await api("/api/pedidos"); const lista = $("#lista-pedidos"); lista.replaceChildren();
  if (!pedidos.length) { lista.append(texto("p", "Aún no se han confirmado pedidos.")); return; }
  pedidos.slice().reverse().forEach((orden) => {
    const fila = document.createElement("div"); fila.className = "record";
    const detalle = document.createElement("div"); detalle.append(texto("strong", `${orden.id} · ${orden.estado}`), texto("small", `${orden.cliente.nombre} · ${new Date(orden.fecha).toLocaleString("es-EC")}`));
    fila.append(detalle, texto("strong", dinero(orden.total))); lista.append(fila);
  });
}

async function cargarBajoStock() {
  const productos = await api("/api/productos/bajo-stock?limite=6"); const lista = $("#bajo-stock"); lista.replaceChildren();
  if (!productos.length) { lista.append(texto("p", "No existen alertas.")); return; }
  productos.forEach((p) => { const fila = document.createElement("div"); fila.className="record"; fila.append(texto("span",p.nombre),texto("strong",`${p.stock} uds.`)); lista.append(fila); });
}

$("#buscar").addEventListener("input", () => cargarProductos().catch((e) => mensaje(e.message,true)));
$("#actualizar-pedidos").addEventListener("click", () => cargarPedidos().catch((e) => mensaje(e.message,true)));
$("#form-cupon").addEventListener("submit", async (evento) => { evento.preventDefault(); try { await api("/api/carrito/cupon", {method:"POST",body:JSON.stringify({codigo:$("#cupon").value})}); mensaje("Cupón aplicado"); await cargarCarrito(); } catch(e) { mensaje(e.message,true); } });
$("#form-compra").addEventListener("submit", async (evento) => {
  evento.preventDefault(); const form = new FormData(evento.currentTarget); const datos = Object.fromEntries(form.entries());
  try {
    const compra = await api("/api/pedidos", {method:"POST", body:JSON.stringify(datos)});
    $("#resultado").hidden = false; $("#resultado-titulo").textContent = `${compra.pedido.id} confirmado · ${dinero(compra.pedido.total)}`;
    $("#resultado-json").textContent = JSON.stringify(compra.factura, null, 2); $("#resultado").scrollIntoView({behavior:"smooth"});
    evento.currentTarget.reset(); mensaje("Compra confirmada y factura simulada generada"); await Promise.all([cargarCarrito(),cargarProductos(),cargarPedidos(),cargarBajoStock(),cargarEstado()]);
  } catch(e) { mensaje(e.message,true); }
});

Promise.all([cargarEstado(), cargarProductos(), cargarCarrito(), cargarPedidos(), cargarBajoStock()]).catch((e) => mensaje(e.message,true));
