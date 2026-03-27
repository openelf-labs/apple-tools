ObjC.import("Foundation");

var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
var params = env ? JSON.parse(env.js) : {};

var safari = Application("Safari");
safari.includeStandardAdditions = true;

var tabIndex = (params.tabIndex !== undefined && params.tabIndex !== null) ? params.tabIndex : -1;
var win;

try {
	win = safari.windows[0];
} catch (e) {
	JSON.stringify({error: "No Safari window open"});
}

var tab;
if (tabIndex >= 0) {
	try {
		tab = win.tabs[tabIndex];
	} catch (e) {
		JSON.stringify({error: "Tab index " + tabIndex + " not found"});
	}
} else {
	tab = win.currentTab();
}

var result = {
	title: "",
	url: "",
	source: ""
};

try {
	result.title = tab.name() || "";
	result.url = tab.url() || "";
	result.source = tab.source() || "";
	// Truncate source to avoid huge output
	if (result.source.length > 50000) {
		result.source = result.source.substring(0, 50000) + "\n... (truncated)";
	}
} catch (e) {
	result.error = "Failed to read tab: " + e.message;
}

JSON.stringify(result);
