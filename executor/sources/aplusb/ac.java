import java.util.*;

public class Main {
	static class AplusB {
		void exec(int a, int b) {
			System.out.println(a + b);
		}
	}
	public static void main(String[] args) {
		AplusB aplusb = new AplusB();
		Scanner in = new Scanner(System.in);
		int a = in.nextInt(), b = in.nextInt();
		aplusb.exec(a, b);
	}	
}
