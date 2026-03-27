ObjC.import("Foundation");

var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
var params = env ? JSON.parse(env.js) : {};

var title = params.title || "Notification";
var message = params.message || "";
var subtitle = params.subtitle || "";
var sound = params.sound !== false;

var app = Application.currentApplication();
app.includeStandardAdditions = true;

var opts = {};
if (subtitle) opts.subtitle = subtitle;
if (sound) opts.soundName = "default";

app.displayNotification(message, Object.assign({withTitle: title}, opts));

JSON.stringify({success: true, title: title, message: message});
