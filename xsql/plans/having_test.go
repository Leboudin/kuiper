package plans

import (
	"errors"
	"fmt"
	"github.com/emqx/kuiper/common"
	"github.com/emqx/kuiper/xsql"
	"github.com/emqx/kuiper/xstream/contexts"
	"reflect"
	"strings"
	"testing"
)

func TestHavingPlan_Apply(t *testing.T) {
	var tests = []struct {
		sql    string
		data   interface{}
		result interface{}
	}{
		{
			sql: `SELECT id1 FROM src1 HAVING avg(id1) > 1`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 2, "f1": "v2"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 5, "f1": "v1"},
						},
					},
				},
			},
			result: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 2, "f1": "v2"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 5, "f1": "v1"},
						},
					},
				},
			},
		},

		{
			sql: `SELECT id1 FROM src1 HAVING sum(id1) > 1`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
			result: nil,
		},

		{
			sql: `SELECT id1 FROM src1 HAVING sum(id1) = 1`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
			result: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
		},

		{
			sql: `SELECT id1 FROM src1 HAVING max(id1) > 10`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
			result: nil,
		},
		{
			sql: `SELECT id1 FROM src1 HAVING max(id1) = 1`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
			result: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						},
					},
				},
			},
		}, {
			sql: "SELECT id1 FROM src1 GROUP BY TUMBLINGWINDOW(ss, 10), f1 having f1 = \"v2\"",
			data: xsql.GroupedTuplesSet{
				{
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 1, "f1": "v1"},
					},
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 3, "f1": "v1"},
					},
				},
				{
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 2, "f1": "v2"},
					},
				},
			},
			result: xsql.GroupedTuplesSet{
				{
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 2, "f1": "v2"},
					},
				},
			},
		}, {
			sql: "SELECT count(*) as c, round(a) as r FROM test Inner Join test1 on test.id = test1.id GROUP BY TumblingWindow(ss, 10), test1.color having a > 100",
			data: xsql.GroupedTuplesSet{
				{
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 1, "a": 122.33}},
							{Emitter: "src2", Message: xsql.Message{"id": 1, "color": "w2"}},
						},
					},
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 5, "a": 177.51}},
							{Emitter: "src2", Message: xsql.Message{"id": 5, "color": "w2"}},
						},
					},
				},
				{
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 2, "a": 89.03}},
							{Emitter: "src2", Message: xsql.Message{"id": 2, "color": "w1"}},
						},
					},
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 4, "a": 14.6}},
							{Emitter: "src2", Message: xsql.Message{"id": 4, "color": "w1"}},
						},
					},
				},
			},
			result: xsql.GroupedTuplesSet{
				{
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 1, "a": 122.33}},
							{Emitter: "src2", Message: xsql.Message{"id": 1, "color": "w2"}},
						},
					},
					&xsql.JoinTuple{
						Tuples: []xsql.Tuple{
							{Emitter: "test", Message: xsql.Message{"id": 5, "a": 177.51}},
							{Emitter: "src2", Message: xsql.Message{"id": 5, "color": "w2"}},
						},
					},
				},
			},
		}, {
			sql: "SELECT * FROM test Inner Join test1 on test.id = test1.id GROUP BY TumblingWindow(ss, 10) having a > 100",
			data: xsql.JoinTupleSets{
				xsql.JoinTuple{
					Tuples: []xsql.Tuple{
						{Emitter: "test", Message: xsql.Message{"id": 1, "a": 122.33}},
						{Emitter: "src2", Message: xsql.Message{"id": 1, "color": "w2"}},
					},
				},
				xsql.JoinTuple{
					Tuples: []xsql.Tuple{
						{Emitter: "test", Message: xsql.Message{"id": 1, "a": 68.55}},
						{Emitter: "src2", Message: xsql.Message{"id": 1, "color": "w2"}},
					},
				},
				xsql.JoinTuple{
					Tuples: []xsql.Tuple{
						{Emitter: "test", Message: xsql.Message{"id": 5, "a": 177.51}},
						{Emitter: "src2", Message: xsql.Message{"id": 5, "color": "w2"}},
					},
				},
			},

			result: xsql.JoinTupleSets{
				xsql.JoinTuple{
					Tuples: []xsql.Tuple{
						{Emitter: "test", Message: xsql.Message{"id": 1, "a": 122.33}},
						{Emitter: "src2", Message: xsql.Message{"id": 1, "color": "w2"}},
					},
				},
				xsql.JoinTuple{
					Tuples: []xsql.Tuple{
						{Emitter: "test", Message: xsql.Message{"id": 5, "a": 177.51}},
						{Emitter: "src2", Message: xsql.Message{"id": 5, "color": "w2"}},
					},
				},
			},
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	contextLogger := common.Log.WithField("rule", "TestHavingPlan_Apply")
	ctx := contexts.WithValue(contexts.Background(), contexts.LoggerKey, contextLogger)
	for i, tt := range tests {
		stmt, err := xsql.NewParser(strings.NewReader(tt.sql)).Parse()
		if err != nil {
			t.Errorf("statement parse error %s", err)
			break
		}

		pp := &HavingPlan{Condition: stmt.Having}
		result := pp.Apply(ctx, tt.data)
		if !reflect.DeepEqual(tt.result, result) {
			t.Errorf("%d. %q\n\nresult mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.sql, tt.result, result)
		}
	}
}

func TestHavingPlanError(t *testing.T) {
	var tests = []struct {
		sql    string
		data   interface{}
		result interface{}
	}{
		{
			sql: `SELECT id1 FROM src1 HAVING avg(id1) > "str"`,
			data: xsql.WindowTuplesSet{
				xsql.WindowTuples{
					Emitter: "src1",
					Tuples: []xsql.Tuple{
						{
							Emitter: "src1",
							Message: xsql.Message{"id1": 1, "f1": "v1"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 2, "f1": "v2"},
						}, {
							Emitter: "src1",
							Message: xsql.Message{"id1": 5, "f1": "v1"},
						},
					},
				},
			},
			result: errors.New("run Having error: invalid operation int64(2) > string(str)"),
		}, {
			sql:    `SELECT id1 FROM src1 HAVING avg(id1) > "str"`,
			data:   errors.New("an error from upstream"),
			result: errors.New("an error from upstream"),
		}, {
			sql: "SELECT id1 FROM src1 GROUP BY TUMBLINGWINDOW(ss, 10), f1 having f1 = \"v2\"",
			data: xsql.GroupedTuplesSet{
				{
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 1, "f1": 3},
					},
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 3, "f1": 3},
					},
				},
				{
					&xsql.Tuple{
						Emitter: "src1",
						Message: xsql.Message{"id1": 2, "f1": "v2"},
					},
				},
			},
			result: errors.New("run Having error: invalid operation int64(3) = string(v2)"),
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	contextLogger := common.Log.WithField("rule", "TestHavingPlan_Apply")
	ctx := contexts.WithValue(contexts.Background(), contexts.LoggerKey, contextLogger)
	for i, tt := range tests {
		stmt, err := xsql.NewParser(strings.NewReader(tt.sql)).Parse()
		if err != nil {
			t.Errorf("statement parse error %s", err)
			break
		}

		pp := &HavingPlan{Condition: stmt.Having}
		result := pp.Apply(ctx, tt.data)
		if !reflect.DeepEqual(tt.result, result) {
			t.Errorf("%d. %q\n\nresult mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.sql, tt.result, result)
		}
	}
}
