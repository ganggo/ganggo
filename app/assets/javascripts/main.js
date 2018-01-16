//= require javascripts/api
//= require javascripts/markdown
//= require javascripts/notification

// change border color if anchor is set
anchors = /#(.+?)$/.exec(window.location.href);
if (anchors !== null) {
  var postElem = $("a[name='" + anchors[1] + "']").closest(".card");
  postElem.addClass("text-white bg-info");
  postElem.find(".card-header").addClass("text-white bg-info");
}

// clicking the cross icon on any kind of
// alert should hide the element
$("#flash-container .alert").click(function() {
  $(this).hide();
});
$("#flash-container .alert-success").fadeOut(2000);
$("#flash-container .alert-danger").fadeOut(5000);

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
      API.posts(postID).likes(positive).post().then(function () {
        var cnt = parseInt($(elem).text());
        $(elem).html(cnt+1);
      });
    });
  });
});
