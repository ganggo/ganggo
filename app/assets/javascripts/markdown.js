// parse all markdown text
$("[data-markdown]").each(function() {
  $(this).html(marked($(this).html()));
});
