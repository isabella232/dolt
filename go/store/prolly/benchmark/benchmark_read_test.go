// Copyright 2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package benchmark

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/dolt/go/store/chunks"
	"github.com/dolthub/dolt/go/store/pool"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/types"
	"github.com/dolthub/dolt/go/store/val"
)

func BenchmarkAll(b *testing.B) {
	benchmarkProllyMap(b, 10_000, 1)
	benchmarkTypesMap(b, 10_000, 1)
	benchmarkProllyMap(b, 100_000, 1)
	benchmarkTypesMap(b, 100_000, 1)
}

func BenchmarkProllySmall(b *testing.B) {
	benchmarkProllyMap(b, 10_000, 10)
}

func BenchmarkTypesSmall(b *testing.B) {
	benchmarkTypesMap(b, 10_000, 10)
}

func BenchmarkProllyMedium(b *testing.B) {
	benchmarkProllyMap(b, 100_000, 1)
}

func BenchmarkTypesMedium(b *testing.B) {
	benchmarkTypesMap(b, 100_000, 1)
}

func BenchmarkProllyLarge(b *testing.B) {
	benchmarkProllyMap(b, 1_000_000, 1)
}

type prollyBench struct {
	m    prolly.Map
	tups [][2]val.Tuple
}

type typesBench struct {
	m    types.Map
	tups [][2]types.Tuple
}

func benchmarkProllyMap(b *testing.B, size, iters uint64) {
	bench := generateProllyBench(size)
	b.ReportAllocs()
	b.ResetTimer()

	b.Run("benchmark prolly map", func(b *testing.B) {
		ctx := context.Background()

		for k := uint64(0); k < iters; k++ {
			for i := 0; i < len(bench.tups); i++ {
				idx := rand.Uint64() % uint64(len(bench.tups))
				key := bench.tups[idx][0]

				_ = bench.m.Get(ctx, key, func(_, _ val.Tuple) (e error) {
					return
				})
			}
			b.ReportAllocs()
		}
	})
}

func benchmarkTypesMap(b *testing.B, size, iters uint64) {
	bench := generateTypesBench(size)
	b.ResetTimer()

	s := strconv.Itoa(int(size))
	b.Run("benchmark types map "+s, func(b *testing.B) {
		for k := uint64(0); k < iters; k++ {
			ctx := context.Background()
			for i := 0; i < len(bench.tups); i++ {
				idx := rand.Uint64() % uint64(len(bench.tups))
				_, _, _ = bench.m.MaybeGet(ctx, bench.tups[idx][0])

				//_, ok, err := bench.m.MaybeGet(ctx, bench.tups[idx][0])
				//assert.NoError(b, err)
				//assert.True(b, ok)
			}
		}
	})
}

func generateProllyBench(size uint64) prollyBench {
	ctx := context.Background()
	ns := newTestNodeStore()

	kd := val.NewTupleDescriptor(
		val.Type{Enc: val.Int64Enc, Nullable: false},
	)
	vd := val.NewTupleDescriptor(
		val.Type{Enc: val.Int64Enc, Nullable: true},
		val.Type{Enc: val.Int64Enc, Nullable: true},
		val.Type{Enc: val.Int64Enc, Nullable: true},
		val.Type{Enc: val.Int64Enc, Nullable: true},
		val.Type{Enc: val.Int64Enc, Nullable: true},
	)

	tups := generateProllyTuples(kd, vd, size)

	tt := make([]val.Tuple, 0, len(tups)*2)
	for i := range tups {
		tt = append(tt, tups[i][0], tups[i][1])
	}

	m, err := prolly.NewMapFromTuples(ctx, ns, kd, vd, tt...)
	if err != nil {
		panic(err)
	}

	return prollyBench{m: m, tups: tups}
}

var shared = pool.NewBuffPool()

func newTestNodeStore() tree.NodeStore {
	ts := &chunks.TestStorage{}
	return tree.NewNodeStore(ts.NewView())
}

func generateProllyTuples(kd, vd val.TupleDesc, size uint64) [][2]val.Tuple {
	src := rand.NewSource(0)

	tups := make([][2]val.Tuple, size)
	kb := val.NewTupleBuilder(kd)
	vb := val.NewTupleBuilder(vd)

	for i := range tups {
		// key
		kb.PutInt64(0, int64(i))
		tups[i][0] = kb.Build(shared)

		// val
		vb.PutInt64(0, src.Int63())
		vb.PutInt64(1, src.Int63())
		vb.PutInt64(2, src.Int63())
		vb.PutInt64(3, src.Int63())
		vb.PutInt64(4, src.Int63())
		tups[i][1] = vb.Build(shared)
	}

	return tups
}

func generateTypesBench(size uint64) typesBench {
	ctx := context.Background()
	tups := generateTypesTuples(size)

	tt := make([]types.Value, len(tups)*2)
	for i := range tups {
		tt[i*2] = tups[i][0]
		tt[(i*2)+1] = tups[i][1]
	}

	m, err := types.NewMap(ctx, newTestVRW(), tt...)
	if err != nil {
		panic(err)
	}

	return typesBench{m: m, tups: tups}
}

func newTestVRW() types.ValueReadWriter {
	ts := &chunks.TestStorage{}
	return types.NewValueStore(ts.NewView())
}

func generateTypesTuples(size uint64) [][2]types.Tuple {
	src := rand.NewSource(0)

	// tags
	t0, t1, t2 := types.Uint(0), types.Uint(1), types.Uint(2)
	t3, t4, t5 := types.Uint(3), types.Uint(4), types.Uint(5)

	tups := make([][2]types.Tuple, size)
	for i := range tups {

		// key
		k := types.Int(i)
		tups[i][0], _ = types.NewTuple(types.Format_Default, t0, k)

		// val
		var vv [5 * 2]types.Value
		for i := 1; i < 10; i += 2 {
			vv[i] = types.Uint(uint64(src.Int63()))
		}
		vv[0], vv[2], vv[4], vv[6], vv[8] = t1, t2, t3, t4, t5

		tups[i][1], _ = types.NewTuple(types.Format_Default, vv[:]...)
	}

	return tups
}
