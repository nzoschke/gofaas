// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

func TestSQL(t *testing.T) {
	suite.Run(t, &sqlTestSuite{
		dbs: map[string]sqlmock.Sqlmock{},
	})
}

type sqlTestSuite struct {
	suite.Suite

	dbs map[string]sqlmock.Sqlmock

	dsn  string
	db   *DB
	mock sqlmock.Sqlmock
}

func (s *sqlTestSuite) mockDB(dsn string) {
	if dsn == "" {
		b := make([]byte, 32)
		rand.Read(b)
		dsn = string(b)
	}

	var err error
	s.dsn = dsn
	if mock, ok := s.dbs[dsn]; ok {
		s.mock = mock
	} else {
		_, s.mock, err = sqlmock.NewWithDSN(dsn)
		s.Require().NoError(err)
		s.dbs[dsn] = s.mock
	}
}

func (s *sqlTestSuite) connect() {
	var err error
	s.db, err = SQL("sqlmock", s.dsn)
	s.Require().NoError(err)
}

func (s *sqlTestSuite) mockPSQL(err error) {
	row := sqlmock.NewRows([]string{"version()", "current_user", "current_database()"}).
		AddRow("test version", "test user", "test database").
		RowError(0, err)
	s.mock.ExpectQuery(`SELECT version\(\), current_user, current_database\(\)`).WillReturnRows(row)
}
func (s *sqlTestSuite) mockMySQL(err error) {
	row := sqlmock.NewRows([]string{"version()", "current_user()", "database()"}).
		AddRow("test version", "test user", "test database").
		RowError(0, err)
	s.mock.ExpectQuery(`SELECT version\(\), current_user\(\), database\(\)`).WillReturnRows(row)
}
func (s *sqlTestSuite) mockMSSQL(err error) {
	row := sqlmock.NewRows([]string{"@@version", "current_user", "db_name()"}).
		AddRow("test version", "test user", "test database").
		RowError(0, err)
	s.mock.ExpectQuery(`SELECT @@version, current_user, db_name\(\)`).WillReturnRows(row)
}
func (s *sqlTestSuite) mockOracle(err error) {
	row := sqlmock.NewRows([]string{"version", "user", "ora_database_name"}).
		AddRow("test version", "test user", "test database").
		RowError(0, err)
	s.mock.ExpectQuery(`SELECT version FROM v\$instance UNION SELECT user, ora_database_name FROM dual`).WillReturnRows(row)
}

func (s *sqlTestSuite) TestPasswordlessURL() {
	s.mockDB("postgres://user@host:port/database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("postgres://user@host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestPasswordURL() {
	s.mockDB("postgres://user:password@host:port/database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("postgres://user@host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestPasswordURLQuery() {
	s.mockDB("postgres://host:port/database?password=password")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("postgres://host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestPasswordURLSchemaless() {
	s.mockDB("user:password@host:port/database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("user@host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestPasswordURLSchemalessUserlessQuery() {
	s.mockDB("host:port/database?password=password")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestWeirdPasswordURL() {
	s.mockDB("user%2Fpassword@host:port/database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("user@host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestWeirderPasswordURL() {
	s.mockDB("user/password@host:port/database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("", s.db.connectionString)
	s.Equal("user@host:port/database", s.db.url)
}

func (s *sqlTestSuite) TestPasswordlessConnectionString() {
	s.mockDB("user=user database=database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("user=user database=database", s.db.connectionString)
	s.Equal("", s.db.url)
}

func (s *sqlTestSuite) TestPasswordConnectionString() {
	s.mockDB("user=user password=password database=database")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("user=user database=database", s.db.connectionString)
	s.Equal("", s.db.url)
}

func (s *sqlTestSuite) TestSemicolonPasswordConnectionString() {
	s.mockDB("odbc:server=localhost;user id=sa;password={foo}};bar};otherthing=thing")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("odbc:server=localhost;user id=sa;otherthing=thing", s.db.connectionString)
	s.Equal("", s.db.url)
}

func (s *sqlTestSuite) TestPSQL() {
	s.mockDB("")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("Postgres", s.db.databaseType)
	s.Equal("test version", s.db.databaseVersion)
	s.Equal("test user", s.db.user)
	s.Equal("test database", s.db.dbname)
}

func (s *sqlTestSuite) TestMySQL() {
	s.mockDB("")
	s.mockPSQL(errors.New(""))
	s.mockMySQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("MySQL", s.db.databaseType)
	s.Equal("test version", s.db.databaseVersion)
	s.Equal("test user", s.db.user)
	s.Equal("test database", s.db.dbname)
}

func (s *sqlTestSuite) TestMSSQL() {
	s.mockDB("")
	s.mockPSQL(errors.New(""))
	s.mockMySQL(errors.New(""))
	s.mockMSSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("MS SQL", s.db.databaseType)
	s.Equal("test version", s.db.databaseVersion)
	s.Equal("test user", s.db.user)
	s.Equal("test database", s.db.dbname)
}

func (s *sqlTestSuite) TestOracle() {
	s.mockDB("")
	s.mockPSQL(errors.New(""))
	s.mockMySQL(errors.New(""))
	s.mockMSSQL(errors.New(""))
	s.mockOracle(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("Oracle", s.db.databaseType)
	s.Equal("test version", s.db.databaseVersion)
	s.Equal("test user", s.db.user)
	s.Equal("test database", s.db.dbname)
}

func (s *sqlTestSuite) TestUnknownDatabase() {
	s.mockDB("")
	s.mockPSQL(errors.New(""))
	s.mockMySQL(errors.New(""))
	s.mockMSSQL(errors.New(""))
	s.mockOracle(errors.New(""))
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	s.Equal("Unknown", s.db.databaseType)
	s.Equal("Unknown", s.db.databaseVersion)
	s.Equal("Unknown", s.db.user)
	s.Equal("Unknown", s.db.dbname)
}

func (s *sqlTestSuite) TestDriverVersionPackage() {
	s.mockDB("")
	s.mockPSQL(nil)
	s.connect()

	s.Require().NoError(s.mock.ExpectationsWereMet())
	//s.Equal("gopkg.in/DATA-DOG/go-sqlmock.v1", s.db.driverVersion)
}
