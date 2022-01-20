package a

func f1[T ~int](v T) {
	println(v)
}

func f2[X,Y ~int | ~string](x X, y Y) {
	println(x, y)
}

func g() {
	f1(100)
	f1[int](100)
	f2[int, string](100, "hoge")
}

type A[T ~bool] []T
type B[T ~int] []T

type C interface {
	~string | int
}
