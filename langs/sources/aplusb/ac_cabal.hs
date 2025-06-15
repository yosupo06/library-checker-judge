-- TODO quickcheck
import Data.Array                    -- array
-- TODO attoparsec
import Data.ByteString               -- bytestring
import Data.Set                      -- container
-- TODO extra
import Control.DeepSeq               -- deeqseq
-- TODO fgl
import Data.Hashable                 -- hashable
-- TODO integer-logarithms
import Control.Lens                  -- lens
import Data.Massiv.Array             -- massiv
import Data.Containers               -- mono-traversable
-- TODO mutable-containers
-- TODO mwc-random
import Control.Monad.State           -- mtl
import Data.Heap                     -- heaps
-- TODO parallel
-- TODO parsec
-- TODO primitive
-- TODO psqueues
import System.Random                 -- random
-- TODO refrection
-- TODO regex-tdfa
-- TODO repa
-- TODO template-haskell
import Data.Text                     -- text
-- TODO tf-random
import Control.Monad.Trans.Class     -- transformers
-- TODO unboxing-vector
-- TODO unordered-containers
-- TODO utility-ht
import Data.Vector                   -- vector
import Data.Vector.Algorithms.Search -- vector-algorithms
-- TODO vector-th-unbox

import Prelude as P

main = do
  inputs <- (P.map P.read . P.words) <$> P.getLine
  P.putStrLn $ show $ P.foldl (+) 0 inputs
