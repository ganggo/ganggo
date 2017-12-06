//= require javascripts/api
//= require javascripts/markdown

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

// handle language switcher events
$("ul li.language").each(function(i, elem) {
  $(elem).click(function() {
    var cookie = "REVEL_LANG=" + $(this).attr("value");
    document.cookie = cookie;
    window.location.reload();
    return false;
  });
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
