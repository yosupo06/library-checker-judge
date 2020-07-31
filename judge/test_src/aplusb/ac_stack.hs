import Data.Array                    -- array
import Control.DeepSeq               -- deeqseq
import Data.Hashable                 -- hashable
import Control.Lens                  -- lens
import Data.Containers               -- mono-traversable
import Control.Monad.State           -- mtl
import System.Random                 -- random
import Data.Strict                   -- strict
import Control.Monad.Trans.Class     -- transformers
import Data.Vector.Algorithms.Search -- vector-algorithms
import Data.Char8                    -- word8

main = do
  inputs <- (map read . words) <$> getLine
  putStrLn $ show $ foldl (+) 0 inputs
