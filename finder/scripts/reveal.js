ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var path = params.path || "";

var finder = Application("Finder");
finder.includeStandardAdditions = true;

finder.reveal(Path(path));
finder.activate();

JSON.stringify({
  success: true,
  path: path,
  message: "Revealed '" + path + "' in Finder.",
});
