sudo apt-get install -y ghc haskell-stack
HOME=/root sudo stack upgrade --binary-only

sudo su -c """
stack install --resolver lts-16.7 array bytestring containers deepseq hashable heaps io-streams lens mutable-containers massiv mono-traversable mtl random strict text transformers vector vector-algorithms word8 &&
cd ~ && stack ghc -- /tmp/haskell_load.hs
""" -- library-checker-user
