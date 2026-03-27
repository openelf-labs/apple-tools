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
  music.soundVolume = params.level;
  JSON.stringify({ success: true, volume: params.level });
}
