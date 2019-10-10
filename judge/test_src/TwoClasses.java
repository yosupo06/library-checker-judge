import java.util.*;

public class TwoClasses {
	static class A {
		void exec() {
			System.out.println("Hello!");
		}
	}
	public static void main(String[] args) {
		A a = new A();
		a.exec();
	}
}
