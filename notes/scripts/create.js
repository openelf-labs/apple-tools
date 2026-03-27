ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Notes");
app.includeStandardAdditions = true;

var title = params.title || "";
var body = params.body || "";
var folderName = params.folder || "OpenELF";

// Find or create the target folder
var targetFolder = null;
var folders = app.folders();
for (var i = 0; i < folders.length; i++) {
  if (folders[i].name() === folderName) {
    targetFolder = folders[i];
    break;
  }
}

if (!targetFolder) {
  targetFolder = app.Folder({ name: folderName });
  app.folders.push(targetFolder);
}

// Create the note with title as name and body as content
var note = app.Note({ name: title, body: body });
targetFolder.notes.push(note);

JSON.stringify({
  success: true,
  name: note.name(),
  folder: targetFolder.name(),
  message: "Note '" + note.name() + "' created in folder '" + targetFolder.name() + "'.",
});
