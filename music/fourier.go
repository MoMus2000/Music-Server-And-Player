package main

import (
	"fmt"
	"math"
	"math/cmplx"
)

func ditfft2(x []float64, y []complex128, n, s int) {
	if n == 1 {
		y[0] = complex(x[0], 0)
		return
	}
	ditfft2(x, y, n/2, 2*s)
	ditfft2(x[s:], y[n/2:], n/2, 2*s)
	for k := 0; k < n/2; k++ {
		tf := cmplx.Rect(1, -2*math.Pi*float64(k)/float64(n)) * y[k+n/2]
		y[k], y[k+n/2] = y[k]+tf, y[k]-tf
	}
}

// Take the wave and split into frequencies
// We do that using a fourier transform
func FourierTransform() {
	const numOfSamples = 8
	var inputBuffer [numOfSamples]float64
	var outputBuffer [numOfSamples]float64

	for i := 0; i < numOfSamples; i++ {
		t := float64(i) / float64(numOfSamples)
		// combining different waves together
		// We have frequency of one and three hertz
		inputBuffer[i] = math.Sin(2*math.Pi*t) + math.Sin(2*math.Pi*t*3) + math.Cos(2*math.Pi*t*5)
	}

	// The question is now how do we decompose / uncompress these waves down
	// to their frequencies that they are made up of?

	// If you take your samples and multiply by waves of certain frequencies
	// If the wave in question is present, the area under the curve is > 0.
	// If not part of signal, the aoc < 0
	// The bigger the aoc, the larger the presence of that signal in the wave.

	// If the signal was shifted it won't work. Thats why you have to try for
	// Both Sin and Cosine waves

	// In the output, we loop through 0 to n hertz.
	// In the output the frequency won't be higher than n hertz anyway.
	// That is a general property of the signal.

	// n samples => n frequencies
	for f := 0; f < numOfSamples; f++ {
		outputBuffer[f] = 0
		for i := 0; i < numOfSamples; i++ {
			t := float64(i) / float64(numOfSamples)
			outputBuffer[f] += inputBuffer[i] * math.Sin(2*math.Pi*float64(f)*t)
		}
		if int(outputBuffer[f]) > 0 {
			fmt.Printf("Frequency %d is part of the signal since AOC is %f\n", f, outputBuffer[f])
		}
	}

	// We mixed 1 hertz and 3 hertz of frequency and that is what
	// we should see in the print statement

	// If we'd replaced sin with cosine we'd see shit
	var outputBufferSin [numOfSamples]float64
	var outputBufferCos [numOfSamples]float64
	var outputBufferEuler [numOfSamples]complex128

	for f := 0; f < numOfSamples; f++ {
		outputBufferSin[f] = 0
		outputBufferCos[f] = 0
		for i := 0; i < numOfSamples; i++ {
			t := float64(i) / float64(numOfSamples)
			outputBufferSin[f] += inputBuffer[i] * math.Sin(2*math.Pi*float64(f)*t)
			outputBufferCos[f] += inputBuffer[i] * math.Cos(2*math.Pi*float64(f)*t)

			// There is a trick to keep track of both sin and cosine
			// The trick is to use euler's number
			// e^(ix) = cos(x) + i*sin(x)
			outputBufferEuler[f] += complex(inputBuffer[i], 0) * cmplx.Rect(1, 2*math.Pi*float64(f)*float64(t))
		}

		fmt.Printf("Using Old Way: %d: Sin: %f Cos: %f \n", f, outputBufferSin[f], outputBufferCos[f])
		fmt.Printf("Using the Euler Way: %d: Sin: %f Cos: %f\n", f, imag(outputBufferEuler[f]), real(outputBufferEuler[f]))

		// The problem with this algorithm is that its N^2

		// If we print out all of the fequencies we care about, we see
		// that the values repeat
		// FFT exploits this repition of symmetry

		// 0 1 0 0 | 0 1 0 0
		// By this property, the first half is the same as the second half
		// For the even samples

		// For the odd ones
		// They are the negative of each other

		// Which means we can cut the amount of computations to half
	}
}

func EfficientFourier() {
	// We write the fourier transform again, this time optimized by n/2
	const numOfSamples = 8
	var outputBufferEuler [numOfSamples]complex128
	var inputBuffer [numOfSamples]float64

	for i := 0; i < numOfSamples; i++ {
		t := float64(i) / float64(numOfSamples)
		// combining different waves together
		// We have frequency of one and three hertz
		inputBuffer[i] = math.Sin(2*math.Pi*t) + math.Sin(2*math.Pi*t*3) + math.Cos(2*math.Pi*t*5)
	}

	for f := 0; f < numOfSamples/2; f++ {
		outputBufferEuler[f] = 0
		outputBufferEuler[f+numOfSamples/2] = 0

		// even
		for i := 0; i < numOfSamples; i += 2 {
			t := float64(i) / numOfSamples
			v := complex(inputBuffer[i], 0) * cmplx.Rect(1, 2*math.Pi*float64(f)*float64(t))
			outputBufferEuler[f] += v
			outputBufferEuler[f+numOfSamples/2] += v
		}

		// odd
		for i := 1; i < numOfSamples; i += 2 {
			t := float64(i) / numOfSamples
			v := complex(inputBuffer[i], 0) * cmplx.Rect(1, 2*math.Pi*float64(f)*float64(t))
			outputBufferEuler[f] += v
			outputBufferEuler[f+numOfSamples/2] -= v
		}
	}

	for i := 0; i < numOfSamples; i++ {
		fmt.Printf("Using the Efficient Way: %d: Sin: %f Cos: %f\n", i, imag(outputBufferEuler[i]), real(outputBufferEuler[i]))
	}

	// This symmetry repeats on split even and odd pairs
	// Hence this efficient algo can be converted into a divide and conquer
	// algorithm, converting the entire thing to Nlog(N)
}

func RunFFT() {
	const numOfSamples = 8
	var inputBuffer [numOfSamples]float64
	for i := 0; i < numOfSamples; i++ {
		t := float64(i) / float64(numOfSamples)
		// combining different waves together
		// We have frequency of one and three hertz
		inputBuffer[i] = math.Sin(2*math.Pi*t) + math.Sin(2*math.Pi*t*3) + math.Cos(2*math.Pi*t*5)
	}
	y := make([]complex128, len(inputBuffer))
	ditfft2(inputBuffer[:], y, len(inputBuffer), 1)
	for _, c := range y {
		fmt.Printf("%8.4f\n", c)
		fmt.Printf("Sin: %f and Cos: %f \n", imag(c), real(c))
	}
}

func main() {
	// EfficientFourier()
	RunFFT()
}
