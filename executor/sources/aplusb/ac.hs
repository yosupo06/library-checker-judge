main = do
  inputs <- (map read . words) <$> getLine
  putStrLn $ show $ foldl (+) 0 inputs
