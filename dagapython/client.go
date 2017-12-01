package dagapython

import (
	"crypto/sha512"
	"fmt"
	"io"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/random"
)

/*Client is used to store the client's private key and index.
All the client's methods are attached to it*/
type Client struct {
	Private abstract.Scalar
	index   int
}

/*ClientMessage stores an authentication request message sent by the client to an arbitrarily chosen server*/
type ClientMessage struct {
	context ContextEd25519
	S       []abstract.Point
	T0      abstract.Point
	proof   ClientProof
}

/*ClientProof stores the client's proof of his computations*/
type ClientProof struct {
	cs abstract.Scalar
	t  []abstract.Point
	c  []abstract.Scalar
	r  []abstract.Scalar
}

/*CreateRequest generates the elements for the authentication request (T0, S) and the generation of the client's proof(s)*/
func (client *Client) CreateRequest(context ContextEd25519) (T0 abstract.Point, S []abstract.Point, s abstract.Scalar, err error) {
	//Step 1: generate ephemeral DH keys
	z := context.C.Scalar().Pick(random.Stream)
	Z := context.C.Point().Mul(nil, z)

	//Step 2: Generate shared secrets with the servers
	shared := make([][]byte, len(context.G.Y))
	for i := 0; i < len(context.G.Y); i++ {
		hasher := sha512.New()
		var writer io.Writer = hasher
		_, err := context.C.Point().Mul(context.G.Y[i], z).MarshalTo(writer)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error in shared secrets: %s", err)
		}
		hash := hasher.Sum(nil)
		shared[i] = hash[:]
	}

	//Step 3: initial linkage tag and commitments
	//Computes the value of the exponent for the initial linkage tag
	exp := context.C.Scalar().One()
	for i := 0; i < len(context.G.Y); i++ {
		exp.Mul(exp, context.C.Scalar().SetBytes(shared[i]))
	}
	T0 = context.C.Point().Mul(context.H[client.index], exp)

	//Computes the commitments
	S = make([]abstract.Point, len(context.G.Y))
	exp = context.C.Scalar().One()
	for i := 0; i < len(context.G.Y)+1; i++ {
		S[i] = context.C.Point().Mul(nil, exp)
		exp.Mul(exp, context.C.Scalar().SetBytes(shared[i]))
	}
	s = exp

	//Add the client's ephemeral public key to the commitments
	/*Prepend taken from comment at
	https://codingair.wordpress.com/2014/07/18/go-appendprepend-item-into-slice/ */
	S = append(S, nil)
	copy(S[1:], S)
	S[0] = Z

	return T0, S, s, nil
}

//GenerateProofCommitments creates and returns the client's commitments t and the random wieghts w
func (client *Client) GenerateProofCommitments(context ContextEd25519, T0 abstract.Point, s abstract.Scalar) (t []abstract.Point, v, w []abstract.Scalar) {
	//Generates w randomly except for w[client.index] = 0
	w = make([]abstract.Scalar, len(context.H))
	for i := range w {
		w[i] = context.C.Scalar().Pick(random.Stream)
	}
	w[client.index] = context.C.Scalar().Zero()

	//Generates random v (2 per client)
	v = make([]abstract.Scalar, 2*len(context.H))
	for i := 0; i < len(v); i++ {
		v[i] = context.C.Scalar().Pick(random.Stream)
	}

	//Generates the commitments t (3 per clients)
	t = make([]abstract.Point, 3*len(context.H))
	for i := 0; i < len(context.H); i++ {
		a := context.C.Point().Mul(context.H[i], w[i])
		b := context.C.Point().Mul(nil, v[2*i])
		t[3*i] = context.C.Point().Add(a, b)

		Sm := context.C.Point().Mul(nil, s)
		c := context.C.Point().Mul(Sm, w[i])
		d := context.C.Point().Mul(nil, v[2*i+1])
		t[3*i+1] = context.C.Point().Add(c, d)

		e := context.C.Point().Mul(T0, w[i])
		f := context.C.Point().Mul(context.H[i], v[2*i+1])
		t[3*i+2] = context.C.Point().Add(e, f)
	}

	return t, v, w
}

//GenerateProofResponses creates the responses to the challenge cs sent by the servers
func (client *Client) GenerateProofResponses(context ContextEd25519, s, cs abstract.Scalar, w, v []abstract.Scalar) (c, r []abstract.Scalar) {
	//Generates the c array
	copy(c, w)
	sum := context.C.Scalar().Zero()
	for _, i := range w {
		sum = context.C.Scalar().Add(sum, i)
	}
	c[client.index] = context.C.Scalar().Sub(cs, sum)

	//Generates the responses
	copy(r, v)
	a := context.C.Scalar().Mul(c[client.index], client.Private)
	r[2*client.index] = context.C.Scalar().Sub(v[2*client.index], a)

	b := context.C.Scalar().Mul(c[client.index], s)
	r[2*client.index+1] = context.C.Scalar().Sub(v[2*client.index+1], b)

	return c, r
}

/*VerifyClientProof checks the validity of a client's proof*/
func VerifyClientProof(msg ClientMessage) bool {
	n := len(msg.context.G.X)
	if len(msg.proof.c) != n {
		return false
	}
	if len(msg.proof.r) != 2*n {
		return false
	}
	if len(msg.proof.t) != 3*n {
		return false
	}

	//Check the commitments
	for i := 0; i < n; i++ {
		a := msg.context.C.Point().Mul(msg.context.G.X[i], msg.proof.c[i])
		b := msg.context.C.Point().Mul(nil, msg.proof.r[2*i])
		ti0 := msg.context.C.Point().Add(a, b)
		if !ti0.Equal(msg.proof.t[3*i]) {
			return false
		}

		c := msg.context.C.Point().Mul(msg.S[len(msg.S)-1], msg.proof.c[i])
		d := msg.context.C.Point().Mul(nil, msg.proof.r[2*i+1])
		ti10 := msg.context.C.Point().Add(c, d)
		if !ti10.Equal(msg.proof.t[3*i+1]) {
			return false
		}

		e := msg.context.C.Point().Mul(msg.T0, msg.proof.c[i])
		f := msg.context.C.Point().Mul(msg.context.H[i], msg.proof.r[2*i+1])
		ti11 := msg.context.C.Point().Add(e, f)
		if !ti11.Equal(msg.proof.t[3*i+2]) {
			return false
		}
	}

	//Check the challenge
	cs := msg.context.C.Scalar().Zero()
	for _, ci := range msg.proof.c {
		cs = msg.context.C.Scalar().Add(cs, ci)
	}
	if !cs.Equal(msg.proof.cs) {
		return false
	}

	return true
}

/*ValidateClientMessage is an utility function to validate that a client message is correclty formed*/
func ValidateClientMessage(msg ClientMessage) bool {
	//Number of clients
	i := len(msg.context.G.X)
	//Number of servers
	j := len(msg.context.G.Y)
	//A commimtment for each server exists and the second element is the generator S=(Z,g,S1,..,Sj)
	if len(msg.S) != j+2 {
		return false
	}
	if msg.S[0] != msg.context.C.Point().Mul(nil, msg.context.C.Scalar().One()) {
		return false
	}
	//T0 not empty
	if msg.T0 == nil {
		return false
	}
	//Proof fields have the correct size
	if len(msg.proof.c) != j || len(msg.proof.r) != 2*i || len(msg.proof.t) != 3*i || msg.proof.cs == nil {
		return false
	}
	return true
}
