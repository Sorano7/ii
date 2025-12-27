true: a -> b -> a
false: a -> b -> b

pair: a -> b -> f -> f (a) (b)
fst: p -> p (true)
snd: p -> p (false)
swap: p -> pair (p >> snd) (p >> fst)

curry: f -> a -> b -> f ([a, b])
uncurry: f -> p ->  f << fst(p) << snd(p)

if: pair
ifNil: a -> b -> x -> (x = nil) (a) (b)

cons: pair
head: fst
tail: snd

comp: f -> g -> x -> x >> g >> f

and: a -> b -> a (b) (false)
or: a -> b -> a (true) (b)
not: a -> a (false) (true)
xor: a -> b -> a (not (b)) (b)
