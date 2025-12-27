true :: a; b; a
false :: a; b; b

pair :: a; b; f; f (a) (b)
fst :: p; p(true)
snd :: p; p(false)
swap :: p; pair << snd(p) << fst(p)

curry :: f; a; b; f([a, b])
uncurry :: f; p; f (fst(p)) (snd(p))

if :: pair
ite :: x; a; b; x (a) (b)
isNil :: x; x = nil
null :: isNil
ifNull :: x; ite (null(x))

cons :: pair
head :: fst
tail :: snd

fcomp :: f; g; x; x >> g >> f

and :: a; b; a (b) (false)
or :: a; b; a (true) (b)
not :: a; a (false) (true)
xor :: a; b; a (not (b)) (b)

succ :: a; a + 1

add :: a; b; a + b
sub :: a; b; a - b
mul :: a; b; a * b
div :: a; b; a / b

foldr :: f; z; xs; ifNull (xs) (z) <> f (head(xs)) <> foldr (f) (z) (tail(xs))
map :: f; xs; foldr << fcomp (cons) (f) << nil << xs
append :: xs; ys; foldr (cons) (ys) (xs)

sum :: foldr (add) (0)