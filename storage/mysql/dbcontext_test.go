package mysql

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Exec(query string, args ...any) (sql.Result, error) {
	argsSlice := m.Called(query, args)
	return argsSlice.Get(0).(sql.Result), argsSlice.Error(1)
}

func (m *MockTx) Prepare(query string) (*sql.Stmt, error) {
	argsSlice := m.Called(query)
	return argsSlice.Get(0).(*sql.Stmt), argsSlice.Error(1)
}

func (m *MockTx) Query(query string, args ...any) (*sql.Rows, error) {
	argsSlice := m.Called(query, args)
	return argsSlice.Get(0).(*sql.Rows), argsSlice.Error(1)
}

type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) { return 1, nil }
func (m *MockResult) RowsAffected() (int64, error) { return 1, nil }

func TestDBContext_Exec(t *testing.T) {
	mockResult := new(MockResult)

	// Case 1: ExecFn override
	ctx := &DBContext{
		ExecFn: func(query string, args ...any) (sql.Result, error) {
			return mockResult, nil
		},
	}
	res, err := ctx.Exec("fake", 1)
	assert.NoError(t, err)
	assert.Equal(t, mockResult, res)
}

func TestDBContext_exec(t *testing.T) {
	mockResult := new(MockResult)

	t.Run("UsingTx", func(t *testing.T) {
		tx := new(MockTx)
		tx.On("Exec", "q1", mock.Anything).Return(mockResult, nil)
		ctx := &DBContext{Tx: tx}
		res, err := ctx.exec("q1", 1)
		assert.NoError(t, err)
		assert.Equal(t, mockResult, res)
	})

	t.Run("UsingConn", func(t *testing.T) {
		mockConn := new(MockedMySqlDb)
		mockConn.On("Exec", "q2", mock.Anything).Return(mockResult, nil)
		ctx := &DBContext{Conn: mockConn}
		res, err := ctx.exec("q2", 1)
		assert.NoError(t, err)
		assert.Equal(t, mockResult, res)
	})

	t.Run("UsingCluster", func(t *testing.T) {
		SetConnectionConfig("test-cluster-x", &ConnectionConfig{Host: "x", DbName: "y"})
		mockDb := new(MockedMySqlDb)
		mockDb.On("Ping").Return(nil)
		mockDb.On("Exec", "q3", mock.Anything).Return(mockResult, nil)
		mockConnector := new(MockedMySqlDbConnector)
		mockConnector.On("Open", "mysql", mock.Anything).Return(mockDb, nil)

		original := helperMySqlConnector
		helperMySqlConnector = mockConnector
		t.Cleanup(func() {
			helperMySqlConnector = original
		})

		ctx := &DBContext{Cluster: "test-cluster-x"}
		res, err := ctx.exec("q3", 1)
		assert.NoError(t, err)
		assert.Equal(t, mockResult, res)
	})

	t.Run("InvalidCluster", func(t *testing.T) {
		ctx := &DBContext{Cluster: "no-cluster"}
		res, err := ctx.exec("bad", 1)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestDBContext_Prepare(t *testing.T) {
	stmt := new(sql.Stmt)
	ctx := &DBContext{
		PrepareFn: func(query string) (*sql.Stmt, error) {
			return stmt, nil
		},
	}
	s, err := ctx.Prepare("SELECT")
	assert.NoError(t, err)
	assert.Equal(t, stmt, s)
}

func TestDBContext_prepare(t *testing.T) {
	stmt := new(sql.Stmt)

	t.Run("UsingTx", func(t *testing.T) {
		tx := new(MockTx)
		tx.On("Prepare", "q1").Return(stmt, nil)
		ctx := &DBContext{Tx: tx}
		s, err := ctx.prepare("q1")
		assert.NoError(t, err)
		assert.Equal(t, stmt, s)
	})

	t.Run("UsingConn", func(t *testing.T) {
		conn := new(MockedMySqlDb)
		conn.On("Prepare", "q2").Return(stmt, nil)
		ctx := &DBContext{Conn: conn}
		s, err := ctx.prepare("q2")
		assert.NoError(t, err)
		assert.Equal(t, stmt, s)
	})

	t.Run("UsingCluster", func(t *testing.T) {
		SetConnectionConfig("c", &ConnectionConfig{Host: "x", DbName: "y"})
		db := new(MockedMySqlDb)
		db.On("Ping").Return(nil)
		db.On("Prepare", "q3").Return(stmt, nil)
		conn := new(MockedMySqlDbConnector)
		conn.On("Open", "mysql", mock.Anything).Return(db, nil)

		original := helperMySqlConnector
		helperMySqlConnector = conn
		t.Cleanup(func() {
			helperMySqlConnector = original
		})

		ctx := &DBContext{Cluster: "c"}
		s, err := ctx.prepare("q3")
		assert.NoError(t, err)
		assert.Equal(t, stmt, s)
	})

	t.Run("InvalidCluster", func(t *testing.T) {
		ctx := &DBContext{Cluster: "invalid"}
		s, err := ctx.prepare("fail")
		assert.Error(t, err)
		assert.Nil(t, s)
	})
}

func TestDBContext_Query(t *testing.T) {
	rows := new(sql.Rows)
	ctx := &DBContext{
		QueryFn: func(query string, args ...any) (*sql.Rows, error) {
			return rows, nil
		},
	}
	r, err := ctx.Query("Q", 1)
	assert.NoError(t, err)
	assert.Equal(t, rows, r)
}

func TestDBContext_Query_FallbackAndError(t *testing.T) {
	rows := new(sql.Rows)

	t.Run("FallbackToConnQuery", func(t *testing.T) {
		conn := new(MockedMySqlDb)
		conn.On("Query", "fallback-q", mock.Anything).Return(rows, nil)

		ctx := &DBContext{Conn: conn}
		r, err := ctx.Query("fallback-q", 1)
		assert.NoError(t, err)
		assert.Equal(t, rows, r)
	})

	t.Run("QueryFnError", func(t *testing.T) {
		ctx := &DBContext{
			QueryFn: func(query string, args ...any) (*sql.Rows, error) {
				return nil, assert.AnError
			},
		}
		r, err := ctx.Query("q", 1)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}

func TestDBContext_query(t *testing.T) {
	rows := new(sql.Rows)

	t.Run("UsingTx", func(t *testing.T) {
		tx := new(MockTx)
		tx.On("Query", "q11", mock.Anything).Return(rows, nil)
		ctx := &DBContext{Tx: tx}
		r, err := ctx.query("q11", 1)
		assert.NoError(t, err)
		assert.Equal(t, rows, r)
	})

	t.Run("UsingConn", func(t *testing.T) {
		conn := new(MockedMySqlDb)
		conn.On("Query", "q22", mock.Anything).Return(rows, nil)
		ctx := &DBContext{Conn: conn}
		r, err := ctx.query("q22", 1)
		assert.NoError(t, err)
		assert.Equal(t, rows, r)
	})

	t.Run("UsingCluster", func(t *testing.T) {
		SetConnectionConfig("z", &ConnectionConfig{Host: "h", DbName: "db"})
		db := new(MockedMySqlDb)
		db.On("Ping").Return(nil)
		db.On("Query", "q33", mock.Anything).Return(rows, nil)
		c := new(MockedMySqlDbConnector)
		c.On("Open", "mysql", mock.Anything).Return(db, nil)

		original := helperMySqlConnector
		helperMySqlConnector = c
		t.Cleanup(func() {
			helperMySqlConnector = original
		})

		ctx := &DBContext{Cluster: "z"}
		r, err := ctx.query("q33", 1)
		assert.NoError(t, err)
		assert.Equal(t, rows, r)
	})

	t.Run("InvalidCluster", func(t *testing.T) {
		ctx := &DBContext{Cluster: "bad"}
		r, err := ctx.query("q44", 1)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}
