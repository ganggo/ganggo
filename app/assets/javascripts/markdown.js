(function(){
  $(".markdown").each(function(index) {
    $(this).html(marked($(this).html()));
  });
})();
