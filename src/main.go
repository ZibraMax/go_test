package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/go-p5/p5"
)

// var wg = sync.WaitGroup{}

const (
	ERROR = iota
	EULER
	HEUN
)
const FPS float64 = 120

const (
	width  = 800
	height = 800
	xmax   = 10.0
	xmin   = -10.0
	ymax   = -10.0
	ymin   = 10.0
)

var t float64 = 0.0

type LinealCollider struct {
	x1       [2]float64
	x2       [2]float64
	color    color.RGBA
	ID       int
	normal   [2]float64
	length   float64
	dx       float64
	dy       float64
	width    float64
	callback func(*Object2D, *LinealCollider, float64)
}

func (collider *LinealCollider) display() {
	draw_line(collider.x1[0], collider.x1[1], collider.x2[0], collider.x2[1], collider.color, collider.width)
	// draw_line(collider.x1[0]+collider.dx/2, collider.x1[1]+collider.dy/2, collider.x1[0]+collider.dx/2+collider.normal[0], collider.x1[1]+collider.dy/2+collider.normal[1], collider.color, 0.5)
}

func (collider *LinealCollider) update() {
	collider.display()
}
func newTrigger(x1 [2]float64, x2 [2]float64) *LinealCollider {
	colorc := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	dx := x2[0] - x1[0]
	dy := x2[1] - x1[1]
	leng := math.Sqrt(dx*dx + dy*dy)
	normal := [2]float64{-dy / leng, dx / leng}
	callback := func(object *Object2D, wall *LinealCollider, t float64) {
		d, _ := point_to_line_distance(object.x[0], object.x[1], wall.x1[0], wall.x1[1], wall.x2[0], wall.x2[1])
		if d <= object.r {
			object.color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
		}
	}
	return &LinealCollider{x1: x1, x2: x2, color: colorc, ID: len(triggers), normal: normal, length: leng, dx: dx, dy: dy, width: 5, callback: callback}
}

func newWall(x1 [2]float64, x2 [2]float64) *LinealCollider {
	colorc := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	dx := x2[0] - x1[0]
	dy := x2[1] - x1[1]
	leng := math.Sqrt(dx*dx + dy*dy)
	normal := [2]float64{-dy / leng, dx / leng}
	callback := func(object *Object2D, wall *LinealCollider, t float64) {
		d, dx := point_to_line_distance(object.x[0], object.x[1], wall.x1[0], wall.x1[1], wall.x2[0], wall.x2[1])
		if d <= object.r {
			fact := 2 * (object.v[0]*dx[0]/d + object.v[1]*dx[1]/d)
			r := [2]float64{object.v[0] - fact*dx[0]/d, object.v[1] - fact*dx[1]/d}
			object.v = r
			object.v[0] *= 0.95
			object.v[1] *= 0.95
		}
	}
	return &LinealCollider{x1: x1, x2: x2, color: colorc, ID: len(walls), normal: normal, length: leng, dx: dx, dy: dy, width: 10, callback: callback}
}

type Object2D struct {
	x          [2]float64
	v          [2]float64
	a          [2]float64
	mass       float64
	ID         int
	forces     [](func(*Object2D, []*Object2D, float64) [2]float64)
	integrator int
	color      color.RGBA
	r          float64
	v0         [2]float64
	xh         []([2]float64)
}

func newObject(x [2]float64, v [2]float64, a [2]float64, mass float64, integrator int) *Object2D {
	forces := [](func(*Object2D, []*Object2D, float64) [2]float64){}
	color := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	return &Object2D{x: x, v: v, a: a, ID: len(objects), forces: forces, mass: mass, integrator: integrator, color: color, r: mass / 10, v0: v}
}

func (object *Object2D) addForce(force func(*Object2D, []*Object2D, float64) [2]float64) {
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
		result_forces := force(object, objects, t)
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
			result_forces := force(euler_object, objects, t+dt)
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
	object.xh = append(object.xh, object.x)
	n := 40
	if len(object.xh) > n {
		object.xh = object.xh[1:]
	}

}

func (object *Object2D) collide(colliders []*LinealCollider) {
	for _, collider := range colliders {
		collider.callback(object, collider, t)
	}
}
func (object *Object2D) object_collide(objects []*Object2D) {
	for id, o := range objects {
		if id != object.ID {
			dx := o.x[0] - object.x[0]
			dy := o.x[1] - object.x[1]
			dist := math.Hypot(dx, dy)
			minDist := o.r + object.r
			if dist < minDist {
				m1 := object.mass
				m2 := o.mass
				v1 := object.v
				v2 := o.v
				x1 := object.x
				x2 := o.x
				dot1 := 2 * m2 / (m1 + m2) * ((v1[0]-v2[0])*(x1[0]-x2[0]) + (v1[1]-v2[1])*(x1[1]-x2[1])) / dist / dist
				dot2 := 2 * m1 / (m1 + m2) * ((v2[0]-v1[0])*(x2[0]-x1[0]) + (v2[1]-v1[1])*(x2[1]-x1[1])) / dist / dist
				delta1 := [2]float64{dot1 * (x1[0] - x2[0]), dot1 * (x1[1] - x2[1])}
				delta2 := [2]float64{dot2 * (x2[0] - x1[0]), dot2 * (x2[1] - x1[1])}
				object.v[0] -= delta1[0]
				object.v[1] -= delta1[1]
				o.v[0] -= delta2[0]
				o.v[1] -= delta2[1]
			}
		}
	}
}

func (object *Object2D) display() {
	if len(object.xh) >= 2 {
		for i := 0; i < len(object.xh)-1; i++ {
			x0 := object.xh[i]
			xf := object.xh[i+1]
			// println(i,uint8((float64(i) / 38.0) * 255.0))
			trace_color := color.RGBA{R: 0, G: 0, B: 0, A: uint8((float64(i) / 38.0) * 255.0)}
			draw_line(x0[0], x0[1], xf[0], xf[1], trace_color, 1)
		}
	}
	p5.StrokeWidth(0)
	p5.Fill(object.color)
	draw_ellipse(object.x[0], object.x[1], object.mass/10)
}
func (object *Object2D) update() {
	object.collide(walls)
	object.collide(triggers)
	object.object_collide(objects)
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

func draw_line(x1 float64, y1 float64, x2 float64, y2 float64, color color.RGBA, width float64) {
	p5.Stroke(color)
	p5.StrokeWidth(width)
	p5.Line(x1, y1, x2, y2)
}

func point_to_line_distance(x float64, y float64, x1 float64, y1 float64, x2 float64, y2 float64) (float64, []float64) {

	A := x - x1
	B := y - y1
	C := x2 - x1
	D := y2 - y1

	dot := A*C + B*D
	len_sq := C*C + D*D
	param := -1.0
	if len_sq != 0 {
		param = dot / len_sq
	}
	var xx, yy float64

	if param < 0 {
		xx = x1
		yy = y1
	} else if param > 1 {
		xx = x2
		yy = y2
	} else {
		xx = x1 + param*C
		yy = y1 + param*D
	}

	dx := x - xx
	dy := y - yy
	return math.Sqrt(dx*dx + dy*dy), []float64{-dx, -dy}
}

var dt float64 = 0.0
var objects []*Object2D
var walls []*LinealCollider
var triggers []*LinealCollider

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
func square_Wall(x0 float64, y0 float64, x1 float64, y1 float64) {
	wall := newWall([2]float64{x0, y0}, [2]float64{x1, y0})
	wall.width = 2
	walls = append(walls, wall)

	wall = newWall([2]float64{x1, y0}, [2]float64{x1, y1})
	wall.width = 2
	walls = append(walls, wall)

	wall = newWall([2]float64{x1, y1}, [2]float64{x0, y1})
	wall.width = 2
	walls = append(walls, wall)

	wall = newWall([2]float64{x0, y1}, [2]float64{x0, y0})
	wall.width = 2
	walls = append(walls, wall)
}

func update() {
	draw_text(0.95*xmin, 0.95*ymin, 15, "t="+fmt.Sprintf("%.3f", t))
	draw_text(0.95*xmin, 0.95*ymax, 15, "d="+fmt.Sprintf("%.3f", descanso))
	for _, wall := range walls {
		wall.update()
	}
	for _, trigger := range triggers {
		trigger.update()
	}
	for _, object := range objects {
		object.update()
	}

	event := p5.Event
	if event.Mouse.Pressed {
		if event.Mouse.Buttons.Contain(p5.ButtonLeft) && descanso > 500.0 {
			earth := newObject([2]float64{event.Mouse.Position.X, event.Mouse.Position.Y}, [2]float64{0.0, 0.0}, [2]float64{0.0, 0.0}, 5.0, EULER)
			earth.color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
			earth.addForce(func(object *Object2D, objects []*Object2D, t float64) [2]float64 {
				return [2]float64{0.0, -9.81 * object.mass}
			})
			objects = append(objects, earth)
			descanso = 0.0
		}
	}
	descanso += dt * 1000
}

// var clicking bool = false
var descanso float64 = 500.0

func setup() {
	square_Wall(xmin, ymin, xmax, ymax)
	wall := newWall([2]float64{xmin, ymin}, [2]float64{xmin * 0.2, ymin * 0.2})
	wall.width = 2
	walls = append(walls, wall)

	wall = newTrigger([2]float64{xmax, ymin}, [2]float64{xmax * 0.2, ymin * 0.2})
	wall.width = 2
	triggers = append(triggers, wall)
	p5.Canvas(width, height)
	p5.PhysCanvas(width, height, xmin, xmax, ymin, ymax)
	p5.Background(color.White)
}
func main() {
	println("Iniciando aplicaci??n")
	p5.Run(setup, frameUpdate)
}
