echo 'Install .NET Core'
wget -q https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
apt install -y  ./packages-microsoft-prod.deb
add-apt-repository universe
apt install -y apt-transport-https
apt update
apt install dotnet-sdk-3.1 -y

echo 'Init C# Project'
dirname="/opt"
project_name="C-Sharp"

su -c """
dotnet new console -o /tmp/${project_name} -lang \"C#\" &&
sed -i -e '/<PropertyGroup>/a <AllowUnsafeBlocks>true</AllowUnsafeBlocks>' /tmp/${project_name}/${project_name}.csproj &&
dotnet add /tmp/${project_name} package System.Runtime.CompilerServices.Unsafe -v 4.6.0 &&
dotnet restore /tmp/${project_name} -r ubuntu.18.04-x64
""" -- library-checker-user

cp -r /tmp/${project_name} ${dirname}/${project_name}

echo 'Install Haskell packages'

stack upgrade --binary-only

su -c """
stack install --resolver lts-16.3 array bytestring containers deepseq hashable heaps io-streams lens mutable-containers massiv mono-traversable mtl random strict text transformers vector vector-algorithms word8 &&
cp /var/haskell_load.hs /tmp/haskell_load.hs &&
cd ~ && stack ghc -- /tmp/haskell_load.hs
""" -- library-checker-user
