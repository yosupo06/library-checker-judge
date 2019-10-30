echo 'Install dlang'
wget https://github.com/ldc-developers/ldc/releases/download/v1.15.0/ldc2-1.15.0-linux-x86_64.tar.xz
tar -xf ldc2-1.15.0-linux-x86_64.tar.xz -C /opt
ln -s /opt/ldc2-1.15.0-linux-x86_64/bin/ldc2 /usr/bin/ldc2

echo 'Install Python(PyPy)'
wget https://bitbucket.org/pypy/pypy/downloads/pypy3.6-v7.1.1-linux64.tar.bz2
tar -xf pypy3.6-v7.1.1-linux64.tar.bz2 -C /opt
ln -s /opt/pypy3.6-v7.1.1-linux64/bin/pypy3 /usr/bin/pypy3

echo 'Install .NET Core'
apt-get -y install mono-mcs
# wget -q https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
# apt install -y  ./packages-microsoft-prod.deb
# add-apt-repository universe
# apt install -y apt-transport-https
# apt update
# apt install dotnet-sdk-3.0 -y

# echo 'Init C# Project'
# dirname="/opt"
# project_name="C-Sharp"

# dotnet new console -o ${dirname}/${project_name} -lang "C#"
# sed -i -e '/<PropertyGroup>/a <AllowUnsafeBlocks>true</AllowUnsafeBlocks>' ${dirname}/${project_name}/${project_name}.csproj
# dotnet add ${dirname}/${project_name} package System.Runtime.CompilerServices.Unsafe -v 4.6.0
