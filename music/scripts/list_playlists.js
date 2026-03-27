ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var se = Application("System Events");
var running = se.processes.whose({ name: "Music" })().length > 0;

if (!running) {
  JSON.stringify({ playlists: [], count: 0, message: "Music is not running" });
} else {
  var music = Application("Music");
  var limit = params.limit || 25;
  var allPlaylists = music.playlists();
  var names = [];

  for (var i = 0; i < allPlaylists.length && names.length < limit; i++) {
    try {
      names.push(allPlaylists[i].name());
    } catch (e) {
      // skip inaccessible playlists
    }
  }

  JSON.stringify({ playlists: names, count: names.length });
}
