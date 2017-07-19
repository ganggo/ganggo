$('textarea.compressed').focus(function () {
  $(this).removeClass('compressed');
  $(this).addClass('expanded');
  $(this).focusout(function() {
    $(this).removeClass('expanded');
    $(this).addClass('compressed');
  });
});
