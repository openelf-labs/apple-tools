ObjC.import("Foundation");

var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
var params = env ? JSON.parse(env.js) : {};

var limit = params.limit || 50;

// Use plist to read bookmarks - more reliable than AppleScript
var fm = $.NSFileManager.defaultManager;
var home = $.NSHomeDirectory().js;
var plistPath = home + "/Library/Safari/Bookmarks.plist";

var result = [];

try {
	var data = $.NSData.dataWithContentsOfFile(plistPath);
	if (data && data.js !== undefined) {
		// Fallback: use defaults read
		var app = Application.currentApplication();
		app.includeStandardAdditions = true;
		var raw = app.doShellScript("defaults read ~/Library/Safari/Bookmarks.plist 2>/dev/null | head -500 || echo '{}'");
		// Since plist parsing in JXA is complex, use plutil
		var jsonStr = app.doShellScript("plutil -convert json -o - ~/Library/Safari/Bookmarks.plist 2>/dev/null || echo '{}'");
		var bookmarks = JSON.parse(jsonStr);

		function extractBookmarks(node, folder, results) {
			if (results.length >= limit) return;
			if (!node) return;

			var title = node.Title || node.URIDictionary && node.URIDictionary.title || "";
			var url = node.URLString || "";

			if (url && node.WebBookmarkType === "WebBookmarkTypeLeaf") {
				results.push({
					title: title,
					url: url,
					folder: folder
				});
			}

			if (node.Children) {
				var folderName = node.Title || folder;
				for (var i = 0; i < node.Children.length && results.length < limit; i++) {
					extractBookmarks(node.Children[i], folderName, results);
				}
			}
		}

		extractBookmarks(bookmarks, "", result);
	}
} catch (e) {
	result = [{error: "Failed to read bookmarks: " + e.message}];
}

JSON.stringify(result);
