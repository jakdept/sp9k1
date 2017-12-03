function close() {
  // $(".preview-container").hide();
  $("#preview-container").style.display = "none";
  $(".image-container").removeClass("blur");
}

function open(target) {
  console.log("opening an image");
  $("#preview-pane").src = target.getAttribute("data-original");
  $("#preview-caption").innerHTML = target.alt;
  $("#preview-container").style.display = "flex";
  $(".image-container").addClass("blur");
  // $("#preview-container").show(0);
}

$(document).ready(function () {

  $("#preview-container").onclick = close;
  // imageCards = document.getElementsByClassName("image-card");
  // for (var i = 0; i < imageCards.length; i++) {
  //   imageCards[i].onflick = function() {
  //     console.log("clicked an image")
  //     open(this);
  //   };
  // }

  $(".image-card").on("click", function () {
    console.log("clicked an image");
    open(this);
  });

  // $("#preview-pane").hide(0);
  // $("#preview-container").hide(0);

  $(window).resize(close)
  $(document).keyup(function (e) {
    if (e.keyCode == 27) {
      // keycode 27 is escape
      close();
    }
  });
});

// Returns a function, that, as long as it continues to be invoked, will not
// be triggered. The function will be called after it stops being called for
// N milliseconds. If `immediate` is passed, trigger the function on the
// leading edge, instead of the trailing.
function debounce(func, wait, immediate) {
  var timeout;
  return function () {
    var context = this,
      args = arguments;
    var later = function () {
      timeout = null;
      if (!immediate) func.apply(context, args);
    };
    var callNow = immediate && !timeout;
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    if (callNow) func.apply(context, args);
  };
};

var filterImages = debounce(function () {
  close();
  var filterString = $("#filter").value;
  // show all images
  $("img").each(function( index ) {
    $( this ).show(0);
  });

  // filter to images that match and hide them
  $("img").filter(function( index ) {
    return $( this ).dataset.original.indexOf(filterString) > -1
  }).each(function( index ) {
    $( this ).hide(0);
  });
}, 250);

$("#filter").on("input", function () {
  filterImages();
});