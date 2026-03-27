ObjC.import("Foundation");

var safari = Application("Safari");
safari.includeStandardAdditions = true;

var result = [];
var windows = safari.windows();

for (var wi = 0; wi < windows.length; wi++) {
	var win = windows[wi];
	var tabs;
	try {
		tabs = win.tabs();
	} catch (e) {
		continue;
	}
	for (var ti = 0; ti < tabs.length; ti++) {
		var tab = tabs[ti];
		try {
			result.push({
				windowIndex: wi,
				tabIndex: ti,
				title: tab.name() || "",
				url: tab.url() || ""
			});
		} catch (e) {
			// skip inaccessible tabs
		}
	}
}

JSON.stringify(result);
