// parse all markdown text
$("[data-markdown]").each(function() {
  var html = $(this).html();
  // parse hashtags
  html = html.replace(/#([^#\s<>]{2,})/ig, '[#$1](/tags/$1)');
  // parse mentions
  var mentionRegex = /@\{\s*?([^\s;]*?)[;\s]+?([^@;\s]+?@[^@;\s]+?)\s*?\}/ig;
  var mentionReplacement = '@[$1](/search/$2)';
  if (mentionRegex.exec(html) === null) {
    mentionRegex = /@\{\s*?([^@;\s]+?@[^@;\s]+?)\s*?\}/ig;
    mentionReplacement = '@[$1](/search/$1)';
  }
  html = html.replace(mentionRegex, mentionReplacement);
  // parse markdown
  html = marked(html);
  $(this).html(html);
});
