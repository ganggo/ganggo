Reveal.initialize({
  keyboard: false, // using custom function
  embedded: true,
  width: "90%",
  height: "100%",
  margin: 0,
  minScale: 1,
  maxScale: 1
});

// focus the current slide in case of scrolling
Reveal.addEventListener('slidechanged', function() {
  Reveal.getCurrentSlide().focus();
});

Reveal.addEventListener('ready', function(event) {
  // the first slide is alway the textarea field
  if (Reveal.getSlides().length > 0) {
    Reveal.slide(1, 0);
  }
  // scroll on big text blobs if we reach
  // the end continue with Reveal slides
  $(document).keydown(function(e){
    var elem = $(Reveal.getCurrentSlide());
    var elemHeight = elem[0].scrollHeight - elem[0].offsetHeight;

    var re = /^(.+\/)(\d*)$/i;
    var page = re.exec(window.location.href);

    if (e.keyCode == 37) { // left
      if (Reveal.getIndices().h == 0) {
        if (page && page.length > 2) {
          var pageNumber = parseInt(page[2])
          if (!isNaN(pageNumber) && pageNumber > 1) {
            window.location = page[1] + (pageNumber - 1);
          } else {
            return true;
          }
        } else {
          window.location = window.location.href + "/1";
        }
      } else {
        Reveal.left();
      }
      return false;
    } else if (e.keyCode == 39) { // right
      if (Reveal.getIndices().h == 10) {
        if (page && page.length > 2) {
          var pageNumber = parseInt(page[2])
          if (!isNaN(pageNumber)) {
            window.location = page[1] + (pageNumber + 1);
          } else {
            return true;
          }
        } else {
          window.location = window.location.href + "/2";
        }
      } else {
        Reveal.right();
      }
      return false;
    } else if (elem.scrollTop() == 0 && e.keyCode == 38) {
      Reveal.up();
      return false;
    } else if (elemHeight != elem.scrollTop() && e.keyCode == 40) {
      var scroll = $(elem).scrollTop();

      $(elem).scrollTop(scroll+50);
      return false;
    } else if (elem.scrollTop() != 0 && e.keyCode == 38) {
      var scroll = $(elem).scrollTop();

      $(elem).scrollTop(scroll-50);
      return false;
    } else if (e.keyCode == 40) {
      Reveal.down();
      return false;
    }
  });

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
});

// on click move to first slide
$("#writePost").click(function() {
  Reveal.slide(0,0,0);
});
