package main

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"github.com/go-p5/p5"
)

var wg = sync.WaitGroup{}

const (
	ERROR = iota
	EULER
	HEUN
)
const FPS float64 = 30

var (
	width  = 1500
	height = 500
	xmax   = 10.0
	xmin   = 0.0
	ymax   = 0.0
	ymin   = 20.0
)
var t float64 = 0.0

type Object2D struct {
	x          [2]float64
	v          [2]float64
	a          [2]float64
	mass       float64
	ID         int
	forces     [](func(*Object2D, float64) [2]float64)
	integrator int
	color      color.RGBA
}

func newObject(x [2]float64, v [2]float64, a [2]float64, mass float64, integrator int) *Object2D {
	forces := [](func(*Object2D, float64) [2]float64){}
	color := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	return &Object2D{x: x, v: v, a: a, ID: len(objects), forces: forces, mass: mass, integrator: integrator, color: color}
}

func (object *Object2D) addForce(force func(*Object2D, float64) [2]float64) {
	object.forces = append(object.forces, force)
}

func (object Object2D) print_details() {
	fmt.Printf("ID: %d X = [%.3f,%.3f], V = [%.3f,%.3f], A = [%.3f,%.3f]. dt=%.3f\n", object.ID, object.x[0], object.x[1], object.v[0], object.v[1], object.a[0], object.a[1], dt)
}

func (object *Object2D) move() {
	sumfx := 0.0
	sumfy := 0.0
	sumfxt1 := 0.0
	sumfyt1 := 0.0
	for _, force := range object.forces {
		result_forces := force(object, t)
		sumfx += result_forces[0]
		sumfy += result_forces[1]
	}
	if object.integrator == EULER {
		object.a[0] = sumfx / object.mass
		object.a[1] = sumfy / object.mass
		object.v[0] = object.v[0] + object.a[0]*dt
		object.v[1] = object.v[1] + object.a[1]*dt
		object.x[0] = object.x[0] + object.v[0]*dt
		object.x[1] = object.x[1] + object.v[1]*dt
	} else {
		euler_object := newObject(object.x, object.v, object.a, object.mass, object.integrator)
		euler_object.a[0] = sumfx / euler_object.mass
		euler_object.a[1] = sumfy / euler_object.mass
		euler_object.v[0] = euler_object.v[0] + euler_object.a[0]*dt
		euler_object.v[1] = euler_object.v[1] + euler_object.a[1]*dt
		euler_object.x[0] = euler_object.x[0] + euler_object.v[0]*dt
		euler_object.x[1] = euler_object.x[1] + euler_object.v[1]*dt

		for _, force := range object.forces {
			result_forces := force(euler_object, t+dt)
			sumfxt1 += result_forces[0]
			sumfyt1 += result_forces[1]
		}

		object.a[0] = (sumfx + sumfxt1) / 2 / object.mass
		object.a[1] = (sumfy + sumfyt1) / 2 / object.mass

		k1vx := euler_object.a[0]
		k1vy := euler_object.a[1]
		k1xx := euler_object.v[0]
		k1xy := euler_object.v[1]

		k2vx := sumfxt1 / object.mass
		k2vy := sumfyt1 / object.mass
		k2xx := object.v[0] + (k1vx)*dt
		k2xy := object.v[1] + (k1vy)*dt
		object.v[0] = object.v[0] + (k1vx+k2vx)/2*dt
		object.v[1] = object.v[1] + (k1vy+k2vy)/2*dt

		object.x[0] = object.x[0] + (k1xx+k2xx)/2*dt
		object.x[1] = object.x[1] + (k1xy+k2xy)/2*dt
	}

}
func (object *Object2D) display() {
	p5.Fill(object.color)
	draw_ellipse(object.x[0], object.x[1], 0.03)
}
func (object *Object2D) update() {
	object.move()
	object.display()
	// object.print_details()
	// wg.Done()
}

func draw_ellipse(x float64, y float64, r float64) {
	p5.Ellipse(x, y, 2*r, 2*r)
}

func draw_text(x float64, y float64, size float64, text string) {
	p5.TextSize(size)
	p5.Text(text, x, y)
}

var dt float64 = 0.0
var objects []*Object2D

func frameUpdate() {
	frame_start_time := time.Now()
	update()
	frame_end_time := time.Now()
	dt = float64(frame_end_time.UnixNano())/1000000000.0 - float64(frame_start_time.UnixNano())/1000000000.0
	total_frame_time := 1.0 / FPS * 1000
	remaining_time := total_frame_time - dt/1000.0
	remaining_time = math.Max(0.0, remaining_time)
	time.Sleep(time.Millisecond * time.Duration(int(remaining_time)))
	dt += remaining_time / 1000
	t += dt
}

func update() {
	draw_text(0.5, 0.5, 15, "t="+fmt.Sprintf("%.3f", t))
	for _, object := range objects {
		// wg.Add(1)
		object.update()
	}
	// wg.Wait()
}
func setup() {
	n := 100
	for i := 0; i < n; i++ {
		object := newObject([2]float64{1.2, 5}, [2]float64{5.0, 10.0 * (float64(i) / float64(n))}, [2]float64{0.0, 0.0}, 1.0, HEUN)
		object.addForce(func(object *Object2D, t float64) [2]float64 { return [2]float64{0.0, -9.81 * object.mass} })
		objects = append(objects, object)
		object.color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	}

	p5.Canvas(width, height)
	p5.PhysCanvas(width, height, xmin, xmax, ymin, ymax)
	p5.Background(color.White)
}
func main() {
	println("Iniciando aplicación --- " + time.Now().String())
	p5.Run(setup, frameUpdate)
}
