module Maybe exposing (..)

type Maybe a
    = Just a
    | Nothing


withDefault : Maybe a -> a -> a
withDefault m default =
    case m of
        Just v ->
            v
        
        Nothing ->
            default