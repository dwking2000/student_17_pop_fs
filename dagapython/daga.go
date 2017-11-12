package dagapython

import (
	"crypto/rand"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

/*Parameters is a struct with the parameters for the computations in DAGA*/
/*type Parameters struct {
	P big.Int
	G big.Int
	Q big.Int
}*/

/*Context is a struct with all the context elements to be used in DAGA*/
/*type Context struct {
	param dsa.Parameters
}*/

/*Group contains the list of client's (X) and server's (Y) public keys*/
type Group struct {
	X []abstract.Point
	Y []abstract.Point
}

/*ContextEd25519 holds all the context elements for DAGA with the ed25519 curve
group is the curve
R is the server's commitments
H is the client's per-round generators*/
type ContextEd25519 struct {
	G Group
	R []abstract.Point
	H []abstract.Point
	C ed25519.Curve
}

/*DSASignature contains the hash of the message signed, and the two values for the DSA signature*/
/*type DSASignature struct {
	hash [64]byte
	r    big.Int
	s    big.Int
}*/

/*GetParameters Return P,G and Q parameters for DAGA*/
/*func GetParameters() dsa.Parameters {
	// Here are a 1024-bit safe prime and generator of the (P - 1)/2 order subgroup from the Python code.
	Pvalue := "124325339146889384540494091085456630009856882741872806181731279018491820800119460022367403769795008250021191767583423221479185609066059226301250167164084041279837566626881119772675984258163062926954046545485368458404445166682380071370274810671501916789361956272226105723317679562001235501455748016154806151119"
	Gvalue := "99656004450068572491707650369312821808187082634000238991378622176696343491115105589981816355495019598158936211590631375413874328242985824977217673016350079715590567506898528605283803802106354523568154237112165652810149860207486982093994904778268429329328161591283210109749627870113664380845204583563547255062"
	// Same but 2048 bits.
	//Pvalue := "27927669199745897480192475403549047216554821662794619165264080639365983255644502375016459217363557816327075710968347577873413531747870995166422966792624628587167099967144352300048688249456511457979188066202485263624876864910790966741324232833539527331240187344632772769944302133859635686652694913901774899716176061200063018486234819466278861754046014136602565681682003785393029271730863251114264009567886052085968472025049680504208350286215846649746729345007798729883244031805808718908355554734961706377253224661315024764137163937689747019988805788185103825179357586792027111676659314369039503661013404299535644931247"
	//Gvalue := "23980964643883997791973764119191343485108478268280187273764110016399758928327299356108443446373039955957739682198803936401869820784993184164369945936082779776397447285391214903950439069492767520728971994214558690769765835532813393423832851620613473408517190997814318932462267642122812979432918572951500898997160033670018434913752563529697839833229914557436668909155867344351296361166233129153503324601602877106470219704921071920596364745260042707475245689874318235973443988270541192532588923789966556622843304211912625888536185817353753787604914145051372838321501368259242189807240634790337143910180278807098418984824"

	P := big.NewInt(0)
	if _, ok := P.SetString(Pvalue, 10); ok == false {
		fmt.Printf("error parsing line %#v\n", Pvalue)
	}
	G := big.NewInt(0)
	if _, ok := G.SetString(Gvalue, 10); ok == false {
		fmt.Printf("error parsing line %#v\n", Gvalue)
	}

	Q := big.NewInt(0)
	Q.Div(Q.Sub(P, big.NewInt(1)), big.NewInt(2)) //Order of subgroup G generates, Q = (P-1)/2.
	fmt.Printf("%T\n", Q)
	return dsa.Parameters{P: P, G: G, Q: Q}
}*/

/*GenerateRandomBytes returns securely generated random bytes.
It will return an error if the system's secure random
number generator fails to function correctly, in which
case the caller should not continue.
https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/ */
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

/*RandReader struct used to implement the io.Reader interface for GenerateRandomBytes*/
type RandReader struct{}

/*Read function for GenerateRandomBytes*/
func (r RandReader) Read(b []byte) (int, error) {
	b, err := GenerateRandomBytes(len(b))
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

/*DsaSign computes the DSA signature of a message using a given private key
Usage: DsaSign(private_key, message) big.Int */
/*func DsaSign(priv dsa.PrivateKey, msg []byte) DSASignature {
	msgHash := sha512.Sum512(msg)

	return DSASignature{msgHash, *big.NewInt(0), *big.NewInt(0)}
}*/
