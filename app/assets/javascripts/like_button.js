// find all like buttons and handle events
$("i.fa-thumbs-o-down, i.fa-thumbs-o-up").each(function(i, elem) {
  var postID = $(elem).attr("data-id");
  if (typeof postID === "undefined") {
    return;
  }
  API.posts(postID).likes.get().then(function(likes) {
    var likeCnt = 0;
    var dislikeCnt = 0;
    var length = Object.keys(likes).length;
    for (var i = 0; i < length; i++) {
      if (likes[i].Positive) {
        likeCnt++;
      } else {
        dislikeCnt++;
      }
    }

    // set db count
    if ($(elem).hasClass("fa-thumbs-o-up")) {
      $(elem).html(likeCnt);
    } else {
      $(elem).html(dislikeCnt);
    }

    // register click event
    $(elem).click(function() {
      var positive = false;
      if ($(elem).hasClass("fa-thumbs-o-up")) {
        positive = true;
      }
      // NOTE catch/delete workaround cause of:
      // https://github.com/yui/yuicompressor/issues/47
      API.posts(postID).likes(positive).post()['catch'](function(http) {
        if (!(http.status >= 200 && http.status < 300)) {
          API.posts(postID).likes['delete']().then(function() {
            var cnt = parseInt($(elem).text());
            $(elem).html(cnt-1);
          });
        }
      }).then(function(result) {
        if (result !== undefined) {
          var cnt = parseInt($(elem).text());
          $(elem).html(cnt+1);
        }
      });
    });
  });
});
