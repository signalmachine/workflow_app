export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "app/_app",
	assets: new Set(["favicon.svg"]),
	mimeTypes: {".svg":"image/svg+xml"},
	_: {
		client: {start:"_app/immutable/entry/start.BrVMXrm7.js",app:"_app/immutable/entry/app.C51tYQ9x.js",imports:["_app/immutable/entry/start.BrVMXrm7.js","_app/immutable/chunks/DY-6GXiH.js","_app/immutable/chunks/CVdb1mTO.js","_app/immutable/chunks/BEQvWqFO.js","_app/immutable/entry/app.C51tYQ9x.js","_app/immutable/chunks/CVdb1mTO.js","_app/immutable/chunks/Cu41DqU0.js","_app/immutable/chunks/8hsZ8H25.js","_app/immutable/chunks/BEQvWqFO.js","_app/immutable/chunks/DoPvU_de.js","_app/immutable/chunks/JLLmeEnA.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js'))
		],
		remotes: {
			
		},
		routes: [
			
		],
		prerendered_routes: new Set(["/app/","/app/admin","/app/admin/access","/app/admin/accounting","/app/admin/inventory","/app/admin/parties","/app/inventory","/app/login","/app/operations","/app/review","/app/settings"]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
