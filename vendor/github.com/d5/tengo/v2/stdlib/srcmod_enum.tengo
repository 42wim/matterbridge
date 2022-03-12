is_enumerable := func(x) {
  return is_array(x) || is_map(x) || is_immutable_array(x) || is_immutable_map(x)
}

is_array_like := func(x) {
  return is_array(x) || is_immutable_array(x)
}

export {
  // all returns true if the given function `fn` evaluates to a truthy value on
  // all of the items in `x`. It returns undefined if `x` is not enumerable.
  all: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    for k, v in x {
      if !fn(k, v) { return false }
    }

    return true
  },
  // any returns true if the given function `fn` evaluates to a truthy value on
  // any of the items in `x`. It returns undefined if `x` is not enumerable.
  any: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    for k, v in x {
      if fn(k, v) { return true }
    }

    return false
  },
  // chunk returns an array of elements split into groups the length of size.
  // If `x` can't be split evenly, the final chunk will be the remaining elements.
  // It returns undefined if `x` is not array.
  chunk: func(x, size) {
    if !is_array_like(x) || !size { return undefined }

    numElements := len(x)
    if !numElements { return [] }

    res := []
    idx := 0
    for idx < numElements {
      res = append(res, x[idx:idx+size])
      idx += size
    }

    return res
  },
  // at returns an element at the given index (if `x` is array) or
  // key (if `x` is map). It returns undefined if `x` is not enumerable.
  at: func(x, key) {
    if !is_enumerable(x) { return undefined }

    if is_array_like(x) {
        if !is_int(key) { return undefined }
    } else {
        if !is_string(key) { return undefined }
    }

    return x[key]
  },
  // each iterates over elements of `x` and invokes `fn` for each element. `fn` is
  // invoked with two arguments: `key` and `value`. `key` is an int index
  // if `x` is array. `key` is a string key if `x` is map. It does not iterate
  // and returns undefined if `x` is not enumerable.
  each: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    for k, v in x {
      fn(k, v)
    }
  },
  // filter iterates over elements of `x`, returning an array of all elements `fn`
  // returns truthy for. `fn` is invoked with two arguments: `key` and `value`.
  // `key` is an int index if `x` is array. It returns undefined if `x` is not array.
  filter: func(x, fn) {
    if !is_array_like(x) { return undefined }

    dst := []
    for k, v in x {
      if fn(k, v) { dst = append(dst, v) }
    }

    return dst
  },
  // find iterates over elements of `x`, returning value of the first element `fn`
  // returns truthy for. `fn` is invoked with two arguments: `key` and `value`.
  // `key` is an int index if `x` is array. `key` is a string key if `x` is map.
  // It returns undefined if `x` is not enumerable.
  find: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    for k, v in x {
      if fn(k, v) { return v }
    }
  },
  // find_key iterates over elements of `x`, returning key or index of the first
  // element `fn` returns truthy for. `fn` is invoked with two arguments: `key`
  // and `value`. `key` is an int index if `x` is array. `key` is a string key if
  // `x` is map. It returns undefined if `x` is not enumerable.
  find_key: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    for k, v in x {
      if fn(k, v) { return k }
    }
  },
  // map creates an array of values by running each element in `x` through `fn`.
  // `fn` is invoked with two arguments: `key` and `value`. `key` is an int index
  // if `x` is array. `key` is a string key if `x` is map. It returns undefined
  // if `x` is not enumerable.
  map: func(x, fn) {
    if !is_enumerable(x) { return undefined }

    dst := []
    for k, v in x {
      dst = append(dst, fn(k, v))
    }

    return dst
  },
  // key returns the first argument.
  key: func(k, _) { return k },
  // value returns the second argument.
  value: func(_, v) { return v }
}
