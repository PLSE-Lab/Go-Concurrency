package p

func pred() bool {
  return true
}

func pp(x int) int {
  if x > 2 && pred() {
    return 5
  }

  var b = pred()
  if b {
    return 6
  }
  return 0
}
