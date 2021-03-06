function close() {
  // $(".preview-container").hide();
  document.getElementById("preview-container").style.display = "none";
  $(".image-container").removeClass("blur");
}

function open(target) {
  console.log("opening an image");
  document.getElementById("preview-pane").src = target.getAttribute("data-original");
  document.getElementById("preview-caption").innerHTML = target.alt;
  document.getElementById("preview-container").style.display = "flex";
  $(".image-container").addClass("blur");
  // $("#preview-container").show(0);
}

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
  // function filterImages() {
  var filterString = $(" #filter-search ")[0].value;
  console.log("filtering images by " + filterString);
  close();

  // filter to images that match and hide them and show others
  $(".image-card").each(function (index) {
    if (this.dataset.original.indexOf(filterString) < 0) {
      $(this).hide(0);
    } else {
      $(this).show(0);
    }
  });
}, 250);
// }

$(document).ready(function () {

  $("#preview-container")[0].onclick = close;
  $("#filter-search").keyup(filterImages);

  $(".image-card").on("click", function () {
    console.log("clicked an image");
    open(this);
  });

  $(window).resize(close)
  $(document).keyup(function (e) {
    if (e.keyCode == 27) {
      // keycode 27 is escape
      close();
    }
  });
});