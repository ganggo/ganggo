// on confirm trigger reshare api call
$("i.fa-retweet").click(function() {
  if (!confirm(msg("javascript.retweet_confirm"))) {
    return;
  }

  var postID = $(this).attr("data-id");
  if (typeof postID === "undefined") {
    return;
  }

  API.posts(postID).reshare.post().then(function() {
    notify(msg("javascript.post_retweeted"));
  });

  return false;
});
