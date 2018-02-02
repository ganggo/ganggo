// nsfw filter
$("p[data-markdown]").each(function() {
  var cardText = $(this);
  var cardBody = cardText.closest(".card-body");
  var re = /#nsfw/i;
  var nsfw = re.exec(cardText.text());

  if (nsfw !== null) {
    var div = $("<div class='alert alert-danger'>");
    div.append("<i class='fa fa-exclamation'>&nbsp;");
    div.append(msg("javascript.nsfw_warning"));
    cardBody.prepend(div);
    cardBody.find("a.gallery").hide();
    cardText.hide();

    div.click(function() {
      cardBody.find("a.gallery").show();
      cardText.show();
      div.remove();
    });
  }
});
