import Data.Array                    -- array
import Data.ByteString               -- bytestring
import Data.Set                      -- container
import Control.DeepSeq               -- deeqseq
import Data.Hashable                 -- hashable
import Data.Heap                     -- heaps
import System.IO.Streams             -- io-streams
import Control.Lens                  -- lens
import Data.Massiv.Array             -- massiv
import Data.Containers               -- mono-traversable
import Control.Monad.State           -- mtl
import System.Random                 -- random
import Data.Strict                   -- strict
import Data.Text                     -- text
import Control.Monad.Trans.Class     -- transformers
import Data.Vector                   -- vector
import Data.Vector.Algorithms.Search -- vector-algorithms
import Data.Char8                    -- word8

main = do
  inputs <- (map read . words) <$> getLine
  putStrLn $ show $ foldl (+) 0 inputs
