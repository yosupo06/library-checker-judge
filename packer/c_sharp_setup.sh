wget -q https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O /tmp/packages-microsoft-prod.deb
sudo apt-get install -y /tmp/packages-microsoft-prod.deb
sudo add-apt-repository universe
sudo apt-get install -y apt-transport-https
sudo apt-get update
sudo apt-get install -y dotnet-sdk-3.1

sudo su -c """
dotnet new console -o /tmp/C-Sharp -lang \"C#\" &&
sed -i -e '/<PropertyGroup>/a <AllowUnsafeBlocks>true</AllowUnsafeBlocks>' /tmp/C-Sharp/C-Sharp.csproj &&
dotnet add /tmp/C-Sharp package System.Runtime.CompilerServices.Unsafe -v 4.6.0 &&
dotnet restore /tmp/C-Sharp -r ubuntu.18.04-x64
""" -- library-checker-user

sudo cp -r /tmp/C-Sharp /opt/C-Sharp
