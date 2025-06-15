using System;
using System.Linq;

static class P
{
    static void Main()
    {
        int[] ab = Console.ReadLine().Split().Select(int.Parse).ToArray();
        Console.WriteLine(ab.Sum());
    }
}
