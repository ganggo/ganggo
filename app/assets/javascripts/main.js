//= require javascripts/api
//= require javascripts/parse_time

(function($) {
  var origAppend = $.fn.append;
  $.fn.append = function () {
    return origAppend.apply(this, arguments).trigger("append");
  };
  $("section").on("append", "article", function() {
    $(this).find(".markdown").each(function() {
      $(this).html(marked($(this).html()));
    });
  });
})(jQuery);

// parse all time fields in slides
$("time").each(function(index, elem) {
  var elem = $(elem);
  var ts = elem.attr("datetime").split(/\s/);
  var i = elem.find("i");
  var text = parseTime(ts[0] + " " + ts[1] + "Z");
  elem.text(" " + text + " ago");
  elem.prepend(i);
});

// find all like buttons and handle events
$(".comment-footer i").each(function(i, elem) {
  var postID = $(elem).attr("data-id");
  if (typeof postID === "undefined") {
    return;
  }
  API.posts(postID).likes.get().then(function(likes) {
    var likeCnt = 0;
    var dislikeCnt = 0;
    $.each(likes, function(i, like) {
      if (like.Positive) {
        likeCnt++;
      } else {
        dislikeCnt++;
      }
    });

    // set db count
    if ($(elem).hasClass("like")) {
      $(elem).html(likeCnt);
    } else {
      $(elem).html(dislikeCnt);
    }

    // register click event
    $(elem).click(function() {
      var positive = false;
      if ($(elem).hasClass("like")) {
        positive = true;
      }
      API.posts(postID).likes(positive).post().then(function () {
        var cnt = parseInt($(elem).text());
        $(elem).html(cnt+1);
      });
    });
  });
});

// nsfw filter
$(".comment-post.post p[data-template]").each(function() {
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
