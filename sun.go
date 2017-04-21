
// Package sun returns the altitude of the Sun for any time and location i.e
// longitude and latitude of observer position (geolocation)
// Can be useful for estimating sky brightness around sunrise/sunset.
//
// Refer to http://en.wikipedia.org/wiki/Declination_of_the_Sun using julian date
// Steps:
// 1. Calculate the ecliptic coordinates of the sun
// 2. Calculate equatorial coordinates
// 3. Convert easily to horizontal coordinates
// 4. From which we get the ALTITUDE of the sun as desired
//
// CREDITS:
// Major portion from https://github.com/giraj/sun-altitude.js/
// Licence : none
//
// Some jde and utility portions based on https://github.com/soniakeys/meeus
// License MIT: http://www.opensource.org/licenses/MIT
//
package sun

import (
	"math"
	"time"
)

const axial_tilt float64 = 23.439

// SunAltitude returns the altuide of the Sun above (+ve) or below (-ve) the horizon in degrees
// Any time zone offset in the input parameter is ignored and the time is
// treated as UTC. So time.Now() and time.Now().UTC() will give the same result.
// Location must be specified in decimal degrees for latitude and longitude
// Typical accuracy is around 0.1 degree
//
func SunAltitude(t time.Time, latitude float64, longitude float64) (altitude float64) {
	jd := timeToJD(t)
	jdn := get_jdn(jd)

	l := between(0, 360, 280.460) + 0.9856474*jdn
	g := between(0, 360, 357.528) + 0.9856003*jdn

	ec_long := get_ecliptic_long(l, g)
	r_asc := get_right_ascension(ec_long)

	// make sure right_ascension is in same quadrant as ecliptic_long
	for angleToQuadrant(ec_long) != angleToQuadrant(r_asc) {
		if r_asc < ec_long {
			r_asc += 90
		} else {
			r_asc += -90
		}
	}
	dec := get_declination(ec_long)
	ha := get_hour_angle(jd, longitude, r_asc)
	return angle_asin(angle_sin(latitude)*angle_sin(dec) + angle_cos(latitude)*angle_cos(dec)*angle_cos(ha))
}

func get_ecliptic_long(l float64, g float64) float64 {
	return l + 1.915*angle_sin(g) + 0.02*angle_sin(2.0*g)
}

func get_right_ascension(ecliptic_long float64) float64 {
	return angle_atan(angle_cos(axial_tilt) * angle_tan(ecliptic_long))
}

// +longitude; positive east
func get_hour_angle(jd float64, longitude float64, right_ascension float64) float64 {
	return between(0, 360, get_gst(jd)) + longitude - right_ascension
}

func get_declination(ecliptic_long float64) float64 {
	return angle_asin(angle_sin(axial_tilt) * angle_sin(ecliptic_long))
}

func get_last_jd_midnight(jd float64) float64 {
	if jd >= math.Floor(jd)+0.5 {
		return math.Floor(jd-1) + 0.5
	} else {
		return math.Floor(jd) + 0.5
	}
}

func get_ut_hours(jd float64, last_jd_midnight float64) float64 {
	return 24 * (jd - last_jd_midnight)
}

func get_gst_hours(jdn_midnight float64, ut_hours float64) float64 {
	gmst := 6.697374558 + 0.06570982441908*jdn_midnight + 1.00273790935*ut_hours
	return between(0.0, 24.0, gmst)
}

// http:#aa.usno.navy.mil/faq/docs/GAST.php
// julian date midnight is every .5 (half)
func get_gst(jd float64) float64 {
	// gst : greenwich mean sidereal time
	jdm := get_last_jd_midnight(jd)
	// gst -> local sidereal time done by adding or subtracting local longitude
	// in hours (degrees / 15). If local position is east of greenwich, then
	// add, else subtract.
	// in degrees!
	return 15 * get_gst_hours(get_jdn(jdm), get_ut_hours(jd, jdm))
}

// suppose max - min is the size of interval (one cycle)
func between(min float64, max float64, val float64) float64 {
	for val < min {
		val += max - min
	}
	for max <= val {
		val -= max - min
	}
	return val
}

func angleToQuadrant(angle float64) float64 {
	angle = between(0, 360, angle)
	if angle < 90. {
		return 1.
	}
	if angle < 180. {
		return 2.
	}
	if angle < 270. {
		return 3.
	}
	if angle < 360. {
		return 4.
	}
	return 4.
}

func to_radians(angle float64) float64 {
	return angle * math.Pi / 180.
}

func to_angle(rad float64) float64 {
	return rad * 180. / math.Pi
}

func angle_sin(x float64) float64 {
	return math.Sin(to_radians(x))
}

func angle_cos(x float64) float64 {
	return math.Cos(to_radians(x))
}

func angle_tan(x float64) float64 {
	return math.Tan(to_radians(x))
}

func angle_atan(x float64) float64 {
	return to_angle(math.Atan(x))
}
func angle_asin(x float64) float64 {
	return to_angle(math.Asin(x))
}

func get_jdn(jd float64) float64 {
	return jd - 2451545.0
}

// TimeToJD takes a Go time.Time and returns a JD as float64.
//
// Any time zone offset in the time.Time is ignored and the time is
// treated as UTC.
func timeToJD(t time.Time) float64 {
	ut := t.UTC()
	y, m, _ := ut.Date()
	d := ut.Sub(time.Date(y, m, 0, 0, 0, 0, 0, time.UTC))
	// time.Time is always Gregorian
	return calendarGregorianToJD(y, int(m), float64(d)/float64(24*time.Hour))
}

// CalendarGregorianToJD converts a Gregorian year, month, and day of month
// to Julian day.
//
// Negative years are valid, back to JD 0.  The result is not valid for
// dates before JD 0.
func calendarGregorianToJD(y, m int, d float64) float64 {
	switch m {
	case 1, 2:
		y--
		m += 12
	}
	a := floorDiv(y, 100)
	b := 2 - a + floorDiv(a, 4)
	// (7.1) p. 61
	return float64(floorDiv64(36525*(int64(y+4716)), 100)) +
		float64(floorDiv(306*(m+1), 10)+b) + d - 1524.5
}

// FloorDiv returns the integer floor of the fractional value (x / y).
//
// It uses integer math only, so is more efficient than using floating point
// intermediate values.  This function can be used in many places where INT()
// appears in AA.  As with built in integer division, it panics with y == 0.
func floorDiv(x, y int) (q int) {
	q = x / y
	if (x < 0) != (y < 0) && x%y != 0 {
		q--
	}
	return
}

// FloorDiv64 returns the integer floor of the fractional value (x / y).
//
// It uses integer math only, so is more efficient than using floating point
// intermediate values.  This function can be used in many places where INT()
// appears in AA.  As with built in integer division, it panics with y == 0.
func floorDiv64(x, y int64) (q int64) {
	q = x / y
	if (x < 0) != (y < 0) && x%y != 0 {
		q--
	}
	return
}
