module Dependency exposing ((?), (?:))

(?) : Maybe a -> a -> a
(?) m a =
    Maybe.withDefault a m

infixl 2 ?

(?:) : Maybe a -> a -> a
(?:) m a =
    Maybe.withDefault a m