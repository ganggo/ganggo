// nsfw filter
$(".card-text[data-markdown]").each(function() {
  var elem = $(this);
  var data = elem.html();
  var re = /#nsfw/i;
  var nsfw = re.exec(data);
  if (nsfw !== null) {
    var div = $("<div class='alert alert-danger'>");
    div.append("<i class='fa fa-exclamation'>");
    div.append(" This content is not safe for work and could contain nude pictures!");
    elem.html(div);
    div.click(function() {
      elem.html(data);
    });
  }
});
