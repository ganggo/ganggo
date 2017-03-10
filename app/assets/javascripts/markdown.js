(function(){
  console.debug("Init GangGo JS");
  $(".markdown").each(function(index) {
    $(this).html(marked($(this).html()));
  });
})();
