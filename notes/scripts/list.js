ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var app = Application("Notes");
app.includeStandardAdditions = true;

var limit = params.limit || 50;
var folderFilter = params.folder || "";

var results = [];
var folders = app.folders();

for (var fi = 0; fi < folders.length; fi++) {
  var folder = folders[fi];
  if (folderFilter && folder.name() !== folderFilter) {
    continue;
  }

  var notes = folder.notes();
  for (var ni = 0; ni < notes.length; ni++) {
    if (results.length >= limit) break;

    var note = notes[ni];
    try {
      var body = "";
      try { body = note.plaintext() || ""; } catch (e) {}
      var snippet = body.substring(0, 200);

      results.push({
        name: note.name(),
        folder: folder.name(),
        snippet: snippet,
        creationDate: note.creationDate().toISOString(),
        modificationDate: note.modificationDate().toISOString(),
      });
    } catch (e) {
      // skip malformed notes
    }
  }

  if (results.length >= limit) break;
}

// Sort by modification date descending (most recent first)
results.sort(function (a, b) {
  return new Date(b.modificationDate) - new Date(a.modificationDate);
});

results = results.slice(0, limit);

JSON.stringify(results);
