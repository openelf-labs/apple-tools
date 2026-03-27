ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var se = Application("System Events");
var running = se.processes.whose({ name: "Music" })().length > 0;

if (!running) {
  JSON.stringify({ success: false, message: "Music is not running" });
} else {
  var music = Application("Music");
  var query = params.query;
  var searchType = params.type || "song";

  if (searchType === "playlist") {
    var playlists = music.playlists.whose({ name: { _contains: query } })();
    if (playlists.length === 0) {
      JSON.stringify({ success: false, message: "No playlist found matching '" + query + "'" });
    } else {
      var pl = playlists[0];
      pl.play();
      JSON.stringify({
        success: true,
        playing: true,
        playlist: { name: pl.name(), trackCount: pl.tracks().length },
      });
    }
  } else {
    var tracks = music.tracks.whose({
      _or: [
        { name: { _contains: query } },
        { artist: { _contains: query } },
        { album: { _contains: query } },
      ],
    })();
    if (tracks.length === 0) {
      JSON.stringify({ success: false, message: "No song found matching '" + query + "'" });
    } else {
      var t = tracks[0];
      t.play();
      JSON.stringify({
        success: true,
        playing: true,
        track: { name: t.name(), artist: t.artist(), album: t.album() },
      });
    }
  }
}
