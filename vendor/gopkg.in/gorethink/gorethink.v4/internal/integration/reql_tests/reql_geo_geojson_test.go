// Code generated by gen_tests.py and process_polyglot.py.
// Do not edit this file directly.
// The template for this file is located at:
// ../template.go.tpl
package reql_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	r "gopkg.in/gorethink/gorethink.v4"
	"gopkg.in/gorethink/gorethink.v4/internal/compare"
)

// Test geoJSON conversion
func TestGeoGeojsonSuite(t *testing.T) {
	suite.Run(t, new(GeoGeojsonSuite))
}

type GeoGeojsonSuite struct {
	suite.Suite

	session *r.Session
}

func (suite *GeoGeojsonSuite) SetupTest() {
	suite.T().Log("Setting up GeoGeojsonSuite")
	// Use imports to prevent errors
	_ = time.Time{}
	_ = compare.AnythingIsFine

	session, err := r.Connect(r.ConnectOpts{
		Address: url,
	})
	suite.Require().NoError(err, "Error returned when connecting to server")
	suite.session = session

	r.DBDrop("test").Exec(suite.session)
	err = r.DBCreate("test").Exec(suite.session)
	suite.Require().NoError(err)
	err = r.DB("test").Wait().Exec(suite.session)
	suite.Require().NoError(err)

}

func (suite *GeoGeojsonSuite) TearDownSuite() {
	suite.T().Log("Tearing down GeoGeojsonSuite")

	if suite.session != nil {
		r.DB("rethinkdb").Table("_debug_scratch").Delete().Exec(suite.session)
		r.DBDrop("test").Exec(suite.session)

		suite.session.Close()
	}
}

func (suite *GeoGeojsonSuite) TestCases() {
	suite.T().Log("Running GeoGeojsonSuite: Test geoJSON conversion")

	{
		// geo/geojson.yaml line #4
		/* ({'$reql_type$':'GEOMETRY', 'coordinates':[0, 0], 'type':'Point'}) */
		var expected_ map[interface{}]interface{} = map[interface{}]interface{}{"$reql_type$": "GEOMETRY", "coordinates": []interface{}{0, 0}, "type": "Point"}
		/* r.geojson({'coordinates':[0, 0], 'type':'Point'}) */

		suite.T().Log("About to run line #4: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, 'type': 'Point', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}, "type": "Point"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #4")
	}

	{
		// geo/geojson.yaml line #6
		/* ({'$reql_type$':'GEOMETRY', 'coordinates':[[0,0], [0,1]], 'type':'LineString'}) */
		var expected_ map[interface{}]interface{} = map[interface{}]interface{}{"$reql_type$": "GEOMETRY", "coordinates": []interface{}{[]interface{}{0, 0}, []interface{}{0, 1}}, "type": "LineString"}
		/* r.geojson({'coordinates':[[0,0], [0,1]], 'type':'LineString'}) */

		suite.T().Log("About to run line #6: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{[]interface{}{0, 0}, []interface{}{0, 1}}, 'type': 'LineString', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{[]interface{}{0, 0}, []interface{}{0, 1}}, "type": "LineString"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #6")
	}

	{
		// geo/geojson.yaml line #8
		/* ({'$reql_type$':'GEOMETRY', 'coordinates':[[[0,0], [0,1], [1,0], [0,0]]], 'type':'Polygon'}) */
		var expected_ map[interface{}]interface{} = map[interface{}]interface{}{"$reql_type$": "GEOMETRY", "coordinates": []interface{}{[]interface{}{[]interface{}{0, 0}, []interface{}{0, 1}, []interface{}{1, 0}, []interface{}{0, 0}}}, "type": "Polygon"}
		/* r.geojson({'coordinates':[[[0,0], [0,1], [1,0], [0,0]]], 'type':'Polygon'}) */

		suite.T().Log("About to run line #8: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{[]interface{}{[]interface{}{0, 0}, []interface{}{0, 1}, []interface{}{1, 0}, []interface{}{0, 0}}}, 'type': 'Polygon', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{[]interface{}{[]interface{}{0, 0}, []interface{}{0, 1}, []interface{}{1, 0}, []interface{}{0, 0}}}, "type": "Polygon"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #8")
	}

	{
		// geo/geojson.yaml line #12
		/* err('ReqlQueryLogicError', 'Expected type NUMBER but found ARRAY.', [0]) */
		var expected_ Err = err("ReqlQueryLogicError", "Expected type NUMBER but found ARRAY.")
		/* r.geojson({'coordinates':[[], 0], 'type':'Point'}) */

		suite.T().Log("About to run line #12: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{[]interface{}{}, 0}, 'type': 'Point', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{[]interface{}{}, 0}, "type": "Point"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #12")
	}

	{
		// geo/geojson.yaml line #14
		/* err('ReqlQueryLogicError', 'Expected type ARRAY but found BOOL.', [0]) */
		var expected_ Err = err("ReqlQueryLogicError", "Expected type ARRAY but found BOOL.")
		/* r.geojson({'coordinates':true, 'type':'Point'}) */

		suite.T().Log("About to run line #14: r.GeoJSON(map[interface{}]interface{}{'coordinates': true, 'type': 'Point', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": true, "type": "Point"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #14")
	}

	{
		// geo/geojson.yaml line #16
		/* err('ReqlNonExistenceError', 'No attribute `coordinates` in object:', [0]) */
		var expected_ Err = err("ReqlNonExistenceError", "No attribute `coordinates` in object:")
		/* r.geojson({'type':'Point'}) */

		suite.T().Log("About to run line #16: r.GeoJSON(map[interface{}]interface{}{'type': 'Point', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"type": "Point"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #16")
	}

	{
		// geo/geojson.yaml line #18
		/* err('ReqlNonExistenceError', 'No attribute `type` in object:', [0]) */
		var expected_ Err = err("ReqlNonExistenceError", "No attribute `type` in object:")
		/* r.geojson({'coordinates':[0, 0]}) */

		suite.T().Log("About to run line #18: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #18")
	}

	{
		// geo/geojson.yaml line #20
		/* err('ReqlQueryLogicError', 'Unrecognized GeoJSON type `foo`.', [0]) */
		var expected_ Err = err("ReqlQueryLogicError", "Unrecognized GeoJSON type `foo`.")
		/* r.geojson({'coordinates':[0, 0], 'type':'foo'}) */

		suite.T().Log("About to run line #20: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, 'type': 'foo', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}, "type": "foo"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #20")
	}

	{
		// geo/geojson.yaml line #22
		/* err('ReqlQueryLogicError', 'Unrecognized field `foo` found in geometry object.', [0]) */
		var expected_ Err = err("ReqlQueryLogicError", "Unrecognized field `foo` found in geometry object.")
		/* r.geojson({'coordinates':[0, 0], 'type':'Point', 'foo':'wrong'}) */

		suite.T().Log("About to run line #22: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, 'type': 'Point', 'foo': 'wrong', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}, "type": "Point", "foo": "wrong"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #22")
	}

	{
		// geo/geojson.yaml line #26
		/* ({'$reql_type$':'GEOMETRY', 'coordinates':[0, 0], 'type':'Point', 'crs':null}) */
		var expected_ map[interface{}]interface{} = map[interface{}]interface{}{"$reql_type$": "GEOMETRY", "coordinates": []interface{}{0, 0}, "type": "Point", "crs": nil}
		/* r.geojson({'coordinates':[0, 0], 'type':'Point', 'crs':null}) */

		suite.T().Log("About to run line #26: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, 'type': 'Point', 'crs': nil, })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}, "type": "Point", "crs": nil}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #26")
	}

	{
		// geo/geojson.yaml line #30
		/* err('ReqlQueryLogicError', 'GeoJSON type `MultiPoint` is not supported.', [0]) */
		var expected_ Err = err("ReqlQueryLogicError", "GeoJSON type `MultiPoint` is not supported.")
		/* r.geojson({'coordinates':[0, 0], 'type':'MultiPoint'}) */

		suite.T().Log("About to run line #30: r.GeoJSON(map[interface{}]interface{}{'coordinates': []interface{}{0, 0}, 'type': 'MultiPoint', })")

		runAndAssert(suite.Suite, expected_, r.GeoJSON(map[interface{}]interface{}{"coordinates": []interface{}{0, 0}, "type": "MultiPoint"}), suite.session, r.RunOpts{
			GeometryFormat: "raw",
			GroupFormat:    "map",
		})
		suite.T().Log("Finished running line #30")
	}
}
