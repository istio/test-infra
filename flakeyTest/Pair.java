import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

class Pair<U, V> {
	public U first;
	public V second;
	public Pair(U first, V second) {
		this.first = first;
		this.second = second;
	}

	public U getFirst() {
		return this.first;
	} 

	public V getSecond() {
		return this.second;
	}

	public void setFirst(U newFirst) {
		this.first = newFirst;
	}

	public void setSecond(V newSecond) {
		this.second = newSecond;
	}
}