ObjC.import("AppKit");
ObjC.import("Foundation");

var pb = $.NSPasteboard.generalPasteboard;
var types = pb.types.js.map(function(t) { return t.js; });

var result = {hasImage: false, hasText: false, text: "", imageFormat: ""};

// Check for text
if (types.indexOf("public.utf8-plain-text") >= 0) {
	var text = pb.stringForType($.NSPasteboardTypeString);
	if (text && text.js) {
		result.hasText = true;
		result.text = text.js;
	}
}

// Check for image
if (types.indexOf("public.png") >= 0 || types.indexOf("public.tiff") >= 0) {
	result.hasImage = true;
	result.imageFormat = types.indexOf("public.png") >= 0 ? "png" : "tiff";
}

JSON.stringify(result);
