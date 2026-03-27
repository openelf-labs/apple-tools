ObjC.import("Foundation");

var se = Application("System Events");
var running = se.processes.whose({ name: "Music" })().length > 0;

if (!running) {
  JSON.stringify({ success: false, message: "Music is not running" });
} else {
  var music = Application("Music");
  music.play();
  JSON.stringify({ success: true, message: "Playback started" });
}
