
// Package sun returns the altitude of the Sun for any time and location i.e
// longitude and latitude of observer position (geolocation)
// Can be useful for estimating sky brightness around sunrise/sunset.
//
// Refer to http://en.wikipedia.org/wiki/Declination_ofTheSun using julian date
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

const axialTilt float64 = 23.439

// SunAltitude returns the altuide of the Sun above (+ve) or below (-ve) the horizon in degrees
// Any time zone offset in the input parameter is ignored and the time is
// treated as UTC. So time.Now() and time.Now().UTC() will give the same result.
// Location must be specified in decimal degrees for latitude and longitude
// Typical accuracy is around 0.1 degree
//
func Altitude(t time.Time, latitude float64, longitude float64) (altitude float64) {
	jd := timeToJD(t)
	jdn := getJdn(jd)

	l := between(0, 360, 280.460) + 0.9856474*jdn
	g := between(0, 360, 357.528) + 0.9856003*jdn

	ecLong := getEclipticLong(l, g)
	rAsc := getRightAscension(ecLong)

	// make sure rightAscension is in same quadrant as eclipticLong
	for angleToQuadrant(ecLong) != angleToQuadrant(rAsc) {
		if rAsc < ecLong {
			rAsc += 90
		} else {
			rAsc += -90
		}
	}
	dec := getDeclination(ecLong)
	ha := getHourAngle(jd, longitude, rAsc)
	return angleAsin(angleSin(latitude)*angleSin(dec) + angleCos(latitude)*angleCos(dec)*angleCos(ha))
}

func getEclipticLong(l float64, g float64) float64 {
	return l + 1.915*angleSin(g) + 0.02*angleSin(2.0*g)
}

func getRightAscension(eclipticLong float64) float64 {
	return angleAtan(angleCos(axialTilt) * angleTan(eclipticLong))
}

// +longitude; positive east
func getHourAngle(jd float64, longitude float64, rightAscension float64) float64 {
	return between(0, 360, getGst(jd)) + longitude - rightAscension
}

func getDeclination(eclipticLong float64) float64 {
	return angleAsin(angleSin(axialTilt) * angleSin(eclipticLong))
}

func getLastJdMidnight(jd float64) float64 {
	if jd >= math.Floor(jd)+0.5 {
		return math.Floor(jd-1) + 0.5
	} else {
		return math.Floor(jd) + 0.5
	}
}

func getUtHours(jd float64, lastJdMidnight float64) float64 {
	return 24 * (jd - lastJdMidnight)
}

func getGstHours(jdnMidnight float64, utHours float64) float64 {
	gmst := 6.697374558 + 0.06570982441908*jdnMidnight + 1.00273790935*utHours
	return between(0.0, 24.0, gmst)
}

// http:#aa.usno.navy.mil/faq/docs/GAST.php
// julian date midnight is every .5 (half)
func getGst(jd float64) float64 {
	// gst : greenwich mean sidereal time
	jdm := getLastJdMidnight(jd)
	// gst -> local sidereal time done by adding or subtracting local longitude
	// in hours (degrees / 15). If local position is east of greenwich, then
	// add, else subtract.
	// in degrees!
	return 15 * getGstHours(getJdn(jdm), getUtHours(jd, jdm))
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

func toRadians(angle float64) float64 {
	return angle * math.Pi / 180.
}

func toAngle(rad float64) float64 {
	return rad * 180. / math.Pi
}

func angleSin(x float64) float64 {
	return math.Sin(toRadians(x))
}

func angleCos(x float64) float64 {
	return math.Cos(toRadians(x))
}

func angleTan(x float64) float64 {
	return math.Tan(toRadians(x))
}

func angleAtan(x float64) float64 {
	return toAngle(math.Atan(x))
}
func angleAsin(x float64) float64 {
	return toAngle(math.Asin(x))
}

func getJdn(jd float64) float64 {
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
