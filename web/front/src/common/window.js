export function scrollTo(offset, callback) {
  const fixedOffset = offset.toFixed();
  const onScroll = function () {
    if (window.pageYOffset.toFixed() === fixedOffset) {
      window.removeEventListener('scroll', onScroll);
      callback();
    }
  };

  window.addEventListener('scroll', onScroll);
  onScroll();
  window.scrollTo({
    top: offset,
    left: 0,
    behavior: 'smooth',
  });
}
