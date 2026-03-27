ObjC.import("Foundation");

var se = Application("System Events");
var running = se.processes.whose({ name: "Music" })().length > 0;

if (!running) {
  JSON.stringify({ playing: false, message: "Music is not running" });
} else {
  var music = Application("Music");
  var state = music.playerState();

  var stateStr = "stopped";
  if (state === "playing") stateStr = "playing";
  else if (state === "paused") stateStr = "paused";

  if (stateStr === "stopped") {
    JSON.stringify({ playing: false, state: stateStr, message: "No track is currently playing" });
  } else {
    var track = music.currentTrack;
    var duration = 0;
    try { duration = track.duration(); } catch (e) {}
    var position = 0;
    try { position = music.playerPosition(); } catch (e) {}

    JSON.stringify({
      playing: stateStr === "playing",
      track: {
        name: track.name(),
        artist: track.artist(),
        album: track.album(),
        duration: Math.round(duration),
        position: Math.round(position),
      },
      state: stateStr,
    });
  }
}
