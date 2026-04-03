

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/(app)/_layout.svelte.js')).default;
export const universal = {
  "ssr": false,
  "prerender": true,
  "load": null
};
export const universal_id = "src/routes/(app)/+layout.ts";
export const imports = ["_app/immutable/nodes/2.B0zyY8HW.js","_app/immutable/chunks/DtQlNIYo.js","_app/immutable/chunks/DY-6GXiH.js","_app/immutable/chunks/CVdb1mTO.js","_app/immutable/chunks/BEQvWqFO.js","_app/immutable/chunks/8hsZ8H25.js","_app/immutable/chunks/Cu41DqU0.js","_app/immutable/chunks/DoPvU_de.js","_app/immutable/chunks/JLLmeEnA.js","_app/immutable/chunks/BuIH40_t.js","_app/immutable/chunks/DLdQXZZv.js","_app/immutable/chunks/TLD4X0FI.js","_app/immutable/chunks/DpxeZ2VQ.js"];
export const stylesheets = ["_app/immutable/assets/FlashBanner.CzYnO2JF.css","_app/immutable/assets/2.BOnfqsss.css"];
export const fonts = [];
