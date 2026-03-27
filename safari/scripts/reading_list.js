ObjC.import("Foundation");

var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
var params = env ? JSON.parse(env.js) : {};

var limit = params.limit || 50;
var result = [];

try {
	var app = Application.currentApplication();
	app.includeStandardAdditions = true;

	var jsonStr = app.doShellScript("plutil -convert json -o - ~/Library/Safari/Bookmarks.plist 2>/dev/null || echo '{}'");
	var bookmarks = JSON.parse(jsonStr);

	function findReadingList(node) {
		if (!node) return null;
		if (node.Title === "com.apple.ReadingList") return node;
		if (node.Children) {
			for (var i = 0; i < node.Children.length; i++) {
				var found = findReadingList(node.Children[i]);
				if (found) return found;
			}
		}
		return null;
	}

	var rlNode = findReadingList(bookmarks);
	if (rlNode && rlNode.Children) {
		for (var i = 0; i < rlNode.Children.length && result.length < limit; i++) {
			var item = rlNode.Children[i];
			var readingListData = item.ReadingList || {};
			result.push({
				title: (item.URIDictionary && item.URIDictionary.title) || item.Title || "",
				url: item.URLString || "",
				preview: readingListData.PreviewText || "",
				dateAdded: readingListData.DateAdded || "",
				dateLastFetched: readingListData.DateLastFetched || ""
			});
		}
	}
} catch (e) {
	result = [{error: "Failed to read reading list: " + e.message}];
}

JSON.stringify(result);
