module Main exposing (..)

import Internal.Dependency exposing (maybeStr)
import Dependency exposing ((?), (?:))


main : Program String
main = 
    maybeStr ? "hello" ?: "hello world"