package xiter

import (
	"cmp"
	"iter"
)

// 値からイテレータをつくる。
func Seq[V any](values ...V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// マップからイテレータをつくる。
func Seq2[K comparable, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// ポインタからイテレータをつくる。
func Optional[V any](v *V) iter.Seq[V] {
	return func(yield func(V) bool) {
		if v == nil {
			return
		}

		_ = yield(*v)
	}
}

// マップのキーと値のペア
type Entry[K any, V any] struct {
	Key   K
	Value V
}

// マップからキーと値のペアのイテレータをつくる。
func Entries[K comparable, V any](m map[K]V) iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		for k, v := range m {
			if !yield(Entry[K, V]{Key: k, Value: v}) {
				return
			}
		}
	}
}

// イテレータを配列に変換する。
func ToArray[V any](seq iter.Seq[V]) []V {
	arr := []V{}
	for v := range seq {
		arr = append(arr, v)
	}
	return arr
}

// イテレータをマップに変換する。
func ToMap[K comparable, V any](seq iter.Seq2[K, V]) map[K]V {
	m := map[K]V{}
	for k, v := range seq {
		m[k] = v
	}
	return m
}

// イテレータの末尾に値を追加する。
func Append[V any](seq iter.Seq[V], values ...V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if !yield(v) {
				return
			}
		}

		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// イテレータの末尾に他のイテレータを結合する。
func Concat[V any](seq iter.Seq[V], others ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if !yield(v) {
				return
			}
		}

		for _, seq := range others {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// 指定した位置に値を追加する。
func Insert[V any](seq iter.Seq[V], index int, values ...V) iter.Seq[V] {
	return func(yield func(V) bool) {
		i := 0

		for v := range seq {
			for i >= index && i < index+len(values) {
				if !yield(values[i-index]) {
					return
				}

				i++
			}

			if !yield(v) {
				return
			}

			i++
		}
	}
}

// 指定した位置の値を削除する。
func Remove[V any](seq iter.Seq[V], index int) iter.Seq[V] {
	return func(yield func(V) bool) {
		i := 0

		for v := range seq {
			if i != index {
				if !yield(v) {
					return
				}
			}

			i++
		}
	}
}

// 位置のイテレータを返す。
func Indices[V any](seq iter.Seq[V]) iter.Seq[int] {
	return func(yield func(int) bool) {
		i := 0

		for range seq {
			if !yield(i) {
				return
			}
			i++
		}
	}
}

// 値を変換したイテレータを返す。
func Map[V1 any, V2 any](seq iter.Seq[V1], f func(V1) V2) iter.Seq[V2] {
	return func(yield func(V2) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// 平坦化したイテレータを返す。
func Flatten[V any](nestedSeq iter.Seq[iter.Seq[V]]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for seq := range nestedSeq {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// 値をイテレータに変換し、それらを結合したイテレータを返す。
func FlatMap[V1 any, V2 any](seq iter.Seq[V1], f func(V1) iter.Seq[V2]) iter.Seq[V2] {
	return func(yield func(V2) bool) {
		for v := range seq {
			for v2 := range f(v) {
				if !yield(v2) {
					return
				}
			}
		}
	}
}

// 値を順に演算する。
func Reduce[V any](seq iter.Seq[V], f func(V, V) V) V {
	next, stop := iter.Pull(seq)
	defer stop()

	v, ok := next()
	if ok {
		for {
			v1, ok := next()
			if !ok {
				break
			}
			v = f(v, v1)
		}
	}

	return v
}

// 初期値と値を順に演算する。
func Fold[V1 any, V2 any](seq iter.Seq[V1], initialValue V2, f func(V2, V1) V2) V2 {
	result := initialValue

	for v := range seq {
		result = f(result, v)
	}

	return result
}

// 値の合計を演算する。
func Sum[V cmp.Ordered | ~complex64 | ~complex128](seq iter.Seq[V]) V {
	return Reduce(seq, func(sum V, v V) V {
		return sum + v
	})
}

// 最大の値を返す。
func Max[V cmp.Ordered](seq iter.Seq[V]) V {
	return Reduce(seq, func(v1 V, v2 V) V {
		return max(v1, v2)
	})
}

// 最小の値を返す。
func Min[V cmp.Ordered](seq iter.Seq[V]) V {
	return Reduce(seq, func(v1 V, v2 V) V {
		return min(v1, v2)
	})
}

// 条件を満たす値の数を返す。
func Count[V any](seq iter.Seq[V], f func(V) bool) int {
	return Fold(seq, 0, func(c int, v V) int {
		if f(v) {
			return c + 1
		}
		return c
	})
}

// 条件を満たす最初の値の位置を返す。
func Index[V any](seq iter.Seq[V], f func(V) bool) int {
	i := 0

	for v := range seq {
		if f(v) {
			return i
		}
		i++
	}

	return -1
}

// 条件を満たす最後の値の位置を返す。
func LastIndex[V any](seq iter.Seq[V], f func(V) bool) int {
	i := 0
	index := -1

	for v := range seq {
		if f(v) {
			index = i
		}
		i++
	}

	return index
}

// 条件を満たす値を返す。
func Find[V any](seq iter.Seq[V], f func(V) bool) (V, bool) {
	for v := range seq {
		if f(v) {
			return v, true
		}
	}

	return *new(V), false
}

// ひとつめの old を new で置き換えたイテレータを返す。
func Replace[V comparable](seq iter.Seq[V], old V, new V) iter.Seq[V] {
	done := false

	return Map(seq, func(v V) V {
		if !done && v == old {
			done = true
			return new
		}

		return v
	})
}

// すべての old を new で置き換えたイテレータを返す。
func ReplaceAll[V comparable](seq iter.Seq[V], old V, new V) iter.Seq[V] {
	return Map(seq, func(v V) V {
		if v == old {
			return new
		}

		return v
	})
}

// 条件を満たす値だけのイテレータを返す。
func Filter[V any](seq iter.Seq[V], f func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if !f(v) {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

// 条件を満たす値を変換したイテレータを返す。
func Collect[V1 any, V2 any](seq iter.Seq[V1], f func(V1) (V2, bool)) iter.Seq[V2] {
	return func(yield func(V2) bool) {
		for v := range seq {
			v2, ok := f(v)
			if !ok {
				continue
			}

			if !yield(v2) {
				return
			}
		}
	}
}

// 先頭n個の値のイテレータを返す。
func Take[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	i := 0
	return TakeWhile(seq, func(v V) bool {
		ok := i < n
		i++
		return ok
	})
}

// 条件を満たし続ける先頭の値のイテレータを返す。
func TakeWhile[V any](seq iter.Seq[V], f func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if !f(v) {
				return
			}

			if !yield(v) {
				return
			}
		}
	}
}

// 先頭n個の値を除いたイテレータを返す。
func Drop[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	i := 0
	return DropWhile(seq, func(v V) bool {
		ok := i < n
		i++
		return ok
	})
}

// 条件を満たし続ける先頭の値を除いたイテレータを返す。
func DropWhile[V any](seq iter.Seq[V], f func(V) bool) iter.Seq[V] {
	done := false
	return func(yield func(V) bool) {
		for v := range seq {
			if !done && f(v) {
				continue
			}

			done = true

			if !yield(v) {
				return
			}
		}
	}
}

// 条件を満たす値が存在したら true を返す。
func Any[V any](seq iter.Seq[V], f func(V) bool) bool {
	for v := range seq {
		if f(v) {
			return true
		}
	}

	return false
}

// すべての値が条件を満たせば true を返す。
func All[V any](seq iter.Seq[V], f func(V) bool) bool {
	for v := range seq {
		if !f(v) {
			return false
		}
	}

	return true
}

// 他のイテレータと比較し、すべての値が一致していたら true を返す。
func Equal[V any](seq1 iter.Seq[V], seq2 iter.Seq[V], f func(V, V) bool) bool {
	next1, stop1 := iter.Pull(seq1)
	next2, stop2 := iter.Pull(seq2)
	defer func() {
		stop1()
		stop2()
	}()

	for {
		v1, ok1 := next1()
		v2, ok2 := next2()

		if !(ok1 && ok2) {
			return ok1 == ok2
		}

		if !f(v1, v2) {
			return false
		}
	}
}

// ２つのイテレータの同じ位置の値をペアにしたイテレータを返す。
func Zip[V1 any, V2 any](seq1 iter.Seq[V1], seq2 iter.Seq[V2]) iter.Seq2[V1, V2] {
	return func(yield func(V1, V2) bool) {
		next1, stop1 := iter.Pull(seq1)
		next2, stop2 := iter.Pull(seq2)
		defer func() {
			stop1()
			stop2()
		}()

		for {
			v1, ok := next1()
			if !ok {
				break
			}

			v2, ok := next2()
			if !ok {
				break
			}

			if !yield(v1, v2) {
				return
			}
		}
	}
}

// 値と位置をペアにしたイテレータを返す。
func ZipWithIndex[V any](seq iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0

		for v := range seq {
			if !yield(i, v) {
				return
			}
		}
	}
}

// 条件を満たす配列と満たさない配列を返す。
func Partition[V any](seq iter.Seq[V], f func(V) bool) ([]V, []V) {
	left, right := []V{}, []V{}

	for v := range seq {
		if f(v) {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	return left, right
}

// 条件を満たす値の直前で分割したふたつの配列を返す。
func Split[V any](seq iter.Seq[V], f func(V) bool) ([]V, []V) {
	left, right := []V{}, []V{}
	flg := false

	for v := range seq {
		if flg {
			right = append(right, v)
			continue
		}

		if f(v) {
			flg = true
			right = append(right, v)
			continue
		}

		left = append(left, v)
	}

	return left, right
}

// 条件を満たす値の直後で分割したふたつの配列を返す。
func SplitAfter[V any](seq iter.Seq[V], f func(V) bool) ([]V, []V) {
	left, right := []V{}, []V{}
	flg := false

	for v := range seq {
		if flg {
			right = append(right, v)
			continue
		}

		left = append(left, v)

		if f(v) {
			flg = true
		}
	}

	return left, right
}

// 条件を満たし続ける先頭部分と残りの部分、ふたつの配列を返す。
func Span[V any](seq iter.Seq[V], f func(V) bool) ([]V, []V) {
	left, right := []V{}, []V{}
	flg := false

	for v := range seq {
		if flg {
			right = append(right, v)
			continue
		}

		if !f(v) {
			flg = true
			right = append(right, v)
			continue
		}

		left = append(left, v)
	}

	return left, right
}

// 値ごとに関数の返すキーでグルーピングしたマップを返す。
func GroupBy[K comparable, V any](seq iter.Seq[V], f func(V) K) map[K][]V {
	m := map[K][]V{}

	for v := range seq {
		k := f(v)
		m[k] = append(m[k], v)
	}

	return m
}

// size 個ずつの配列を返すイテレーターを返す。
func Grouped[V any](seq iter.Seq[V], size int) iter.Seq[[]V] {
	if size <= 0 {
		panic("size must be > 0")
	}

	return func(yield func([]V) bool) {
		group := make([]V, 0, size)

		for v := range seq {
			group = append(group, v)

			if len(group) >= size {
				if !yield(group) {
					return
				}

				group = group[:0]
			}
		}

		if len(group) > 0 {
			_ = yield(group)
		}
	}
}

// size 個ずつの配列を step ずつずらして返すイテレーターを返す。
func Sliding[V any](seq iter.Seq[V], size int, step int) iter.Seq[[]V] {
	if size <= 0 || step <= 0 {
		panic("size and step must be > 0")
	}

	if step >= size {
		return func(yield func([]V) bool) {
			window := make([]V, 0, size)
			skip := 0

			for v := range seq {
				if skip > 0 {
					skip--
					continue
				}

				window = append(window, v)

				if len(window) >= size {
					if !yield(window) {
						return
					}

					window = window[:0]
					skip = step - size
				}
			}

			if len(window) > 0 {
				_ = yield(window)
			}
		}
	}

	return func(yield func([]V) bool) {
		window := make([]V, 0, size)

		for v := range seq {
			window = append(window, v)

			if len(window) >= size {
				if !yield(window) {
					return
				}

				window = window[step:]
			}
		}

		if len(window) > 0 {
			_ = yield(window)
		}
	}
}
