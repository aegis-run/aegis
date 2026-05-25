package must

func NotError[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func Cast[T any](v T, ok bool) T {
	if !ok {
		panic("cast failed")
	}

	return v
}
