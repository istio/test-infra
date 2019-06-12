import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.DocumentBuilder;
import org.w3c.dom.Document;
import org.w3c.dom.NodeList;
import org.w3c.dom.Node;
import org.w3c.dom.NamedNodeMap;
import org.w3c.dom.Element;
import java.util.HashMap;
import java.io.FileWriter;
import java.io.BufferedWriter;
import javax.xml.parsers.ParserConfigurationException;
import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerException;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.dom.DOMSource;
import javax.xml.transform.stream.StreamResult;
import org.w3c.dom.Attr;
import java.util.regex.Pattern;
import java.util.Calendar;
import javax.xml.transform.OutputKeys;
import java.nio.file.Paths;
import org.xml.sax.InputSource;
import java.io.StringReader;
import com.google.cloud.storage.Blob;
import com.google.cloud.storage.BlobId;
import com.google.cloud.storage.Storage.BlobListOption;
import com.google.cloud.storage.Storage;
import com.google.cloud.storage.StorageOptions;
import com.google.api.gax.paging.Page;
import com.google.cloud.storage.BlobInfo;
import java.nio.file.Files;
import java.io.StringWriter;
import java.text.SimpleDateFormat;
import java.util.Date;
import static java.nio.charset.StandardCharsets.UTF_8;


/**
 * Read junit.xml files in istio folder which contains the result of test running.
 * Output <output_file_path>.xml into the istioFlakeyTest bucket that contains the number of times 
 * each test suite and test case is run and the number of times it fails in order to calculate the flakeyness of the tests.
 * To edit the folders to read (pre/post submit checks), edit the command.sh in the folder to include the path to the folders in gs://\.
 * The two parameters are optional. If not specified, the program will run the past 7 days of results and output to result.xml\.
 * to read the 
 * If gcloud not installed, run `curl https://sdk.cloud.google.com | bash; exec -l $SHELL; ` to install gcloud to use gsutil.
 * After shell restarts, run `gcloud init`.
 * To avoid the "anomynous user error", run `gcloud auth application-default login`.
 * project name: istioFlakeyTest in gcloud
 * compile: javac -cp ".:jars/*" TotalFlakey.java
 * run: java -cp ".:jars/*" TotalFlakey
 * upload files to google cloud: gsutil cp [LOCAL_OBJECT_LOCATION] gs://[DESTINATION_BUCKET_NAME]/
 * File runs daily and calculate the flakey result for the past 30 and 7 days.
 */
public class TotalFlakey {
	
	static String bucketName = "istio-flakey-test";
	//static String pathToReadInput = "readPastJunitCommand.sh";
	static String pathToReadInput = "testCommand.sh";
	static String dataFolder = "temp";
	static String pathToDeleteTempCommand = "removeTempFolderCommand.sh";

	/*
	 * Add testcase to HashMap when the case is proven to be successful.
	 */
	private static HashMap<String, Pair<Integer, Integer>> addSuccessfulCase(HashMap<String, Pair<Integer, Integer>> caseCollection, String caseName) {
		if (caseCollection.containsKey(caseName)) {
	    	Pair<Integer, Integer> caseResult = caseCollection.get(caseName);
	    	caseResult.setSecond(caseResult.getSecond() + 1);
	    	caseCollection.put(caseName, caseResult);
	    } else {
	    	Pair<Integer, Integer> caseResult = new Pair<Integer, Integer> (0, 1);
	    	caseCollection.put(caseName, caseResult);
	    }
	    return caseCollection;
	}

	/*
	 * Check the number of failures and values in xml elements to determine if the testsuite/testcase failed.
	 */
	public static void identifyFailures(HashMap<String, HashMap<String, Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>>>> fullFlakey, Document doc, String branch) {
		int tests;
		NodeList nodeList = doc.getElementsByTagName("testsuite");
	    for(int x=0,size= nodeList.getLength(); x<size; x++) {
	    	Node curNode = nodeList.item(x);
	    	
	    	if (curNode.getNodeType() == Node.ELEMENT_NODE) {
	    		HashMap<String, Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>>> flakey = new HashMap<>();
    			if (fullFlakey.containsKey(branch)) {
    				flakey = fullFlakey.get(branch);
    			}
	    		if (curNode.hasAttributes()) {
	    			NamedNodeMap nodeMap = curNode.getAttributes();
	    			String suiteName = nodeMap.getNamedItem("name").getNodeValue();
	    			int numSuiteFailures = Integer.parseInt(nodeMap.getNamedItem("failures").getNodeValue());
	    			int numSuiteTests = Integer.parseInt(nodeMap.getNamedItem("tests").getNodeValue());
	    			
 	    			if (flakey.containsKey(suiteName)) {
	    				Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>> result = flakey.get(suiteName);
	    				Pair<Integer, Integer> suiteResult = result.getFirst();
	    				HashMap<String, Pair<Integer, Integer>> caseCollection = result.getSecond();
	    				int suiteTotal = suiteResult.getSecond();
	    				suiteResult.setSecond(suiteTotal + 1);

	    				if (numSuiteFailures != 0) {
							int suiteFailure = suiteResult.getFirst();
							suiteResult.setFirst(suiteFailure + 1);
							if (curNode.hasChildNodes()) {
	    						NodeList childNodes = curNode.getChildNodes();
	    						for (int y = 0; y < childNodes.getLength(); y ++) {
	    							Node testCase = childNodes.item(y);
	    							if (testCase.getNodeType() == Node.ELEMENT_NODE && testCase.getNodeName().equals("testcase")) {
	    								NamedNodeMap casemap = testCase.getAttributes();
	    								
	    								String className = casemap.getNamedItem("classname").getNodeValue();
	    								String method = casemap.getNamedItem("name").getNodeValue();
	    								String caseName = suiteName + "*" + className + "|" + method;
	    								NodeList caseChildren = testCase.getChildNodes();
    									Boolean containsFailure = false;
    									for (int k = 0; k < caseChildren.getLength(); k ++) {
    										Node child = caseChildren.item(k);
    										if (child.getNodeName().equals("failure")) {
    											containsFailure = true;
    											if (caseCollection.containsKey(caseName)) {
    												Pair<Integer, Integer> caseResult = caseCollection.get(caseName);
    												caseResult.setFirst(caseResult.getFirst() + 1);
    												caseResult.setSecond(caseResult.getSecond() + 1);
    											
    												
    												caseCollection.put(caseName, caseResult);
    											} else {
    												Pair<Integer, Integer> caseResult = new Pair<Integer, Integer>(1, 1);
    												
    												caseCollection.put(caseName, caseResult);

    											}
    											break;
    										}
    									}
    									if (containsFailure == false) {
    										caseCollection = addSuccessfulCase(caseCollection, caseName);
    									}
    								}
    							}
    						}
						} else {
							if (curNode.hasChildNodes()) {
	    						NodeList childNodes = curNode.getChildNodes();
	    						for (int y = 0; y < childNodes.getLength(); y ++) {
	    							Node testCase = childNodes.item(y);
	    							if (testCase.getNodeType() == Node.ELEMENT_NODE && testCase.getNodeName().equals("testcase")) {
	    								NamedNodeMap casemap = testCase.getAttributes();
	    								String className = casemap.getNamedItem("classname").getNodeValue();
	    								String method = casemap.getNamedItem("name").getNodeValue();
	    								String caseName = suiteName + "*" + className + "|" + method;
										caseCollection = addSuccessfulCase(caseCollection, caseName);
									}
								}
							}
						}
						result.setFirst(suiteResult);
						result.setSecond(caseCollection);
						flakey.put(suiteName, result);
						fullFlakey.put(branch, flakey);

					} else {
						Pair<Integer, Integer> suiteResult = new Pair<>(0, 1);
						HashMap<String, Pair<Integer, Integer>> caseCollection = new HashMap<>();
						if (numSuiteFailures != 0) {
							int suiteFailure = suiteResult.getFirst();
							suiteResult.setFirst(suiteFailure + 1);
							if (curNode.hasChildNodes()) {
	    						NodeList childNodes = curNode.getChildNodes();
	    						for (int y = 0; y < childNodes.getLength(); y ++) {
	    							Node testCase = childNodes.item(y);
	    							if (testCase.getNodeType() == Node.ELEMENT_NODE && testCase.getNodeName().equals("testcase")) {
	    								NamedNodeMap casemap = testCase.getAttributes();
	    								String className = casemap.getNamedItem("classname").getNodeValue();
	    								String method = casemap.getNamedItem("name").getNodeValue();
	    								String caseName = suiteName + "*" + className + "|" + method;
	    								
	    								NodeList caseChildren = testCase.getChildNodes();
    									Boolean containsFailure = false;
    									for (int k = 0; k < caseChildren.getLength(); k ++) {
    										Node child = caseChildren.item(k);
    										if (child.getNodeName().equals("failure")) {
    											containsFailure = true;
    											Pair<Integer, Integer> caseResult = new Pair<Integer, Integer>(1, 1);
    												
    											caseCollection.put(caseName, caseResult);
    										}
    										break;
    									}
    									if (containsFailure == false) {
    										caseCollection = addSuccessfulCase(caseCollection, caseName);
    									}
    								}
    							}
    						}
    					} else {
							if (curNode.hasChildNodes()) {
	    						NodeList childNodes = curNode.getChildNodes();
	    						for (int y = 0; y < childNodes.getLength(); y ++) {
	    							Node testCase = childNodes.item(y);
	    							if (testCase.getNodeType() == Node.ELEMENT_NODE && testCase.getNodeName().equals("testcase")) {
	    								NamedNodeMap casemap = testCase.getAttributes();
	    								String className = casemap.getNamedItem("classname").getNodeValue();
	    								String method = casemap.getNamedItem("name").getNodeValue();
	    								String caseName = suiteName + "*" + className + "|" + method;
										caseCollection = addSuccessfulCase(caseCollection, caseName);
									}
								}
							}
						}
						
						Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>> result = new Pair<>(suiteResult, caseCollection);
						flakey.put(suiteName, result);
						fullFlakey.put(branch, flakey);
					}
				}
			}
		}
	}

	/*
	 * Convert xml document to String to be written to file.
	 */
	public static String toString(Document doc) {
	    try {
	        StringWriter sw = new StringWriter();
	        TransformerFactory tf = TransformerFactory.newInstance();
	        Transformer transformer = tf.newTransformer();
	        transformer.setOutputProperty(OutputKeys.OMIT_XML_DECLARATION, "no");
	        transformer.setOutputProperty(OutputKeys.METHOD, "xml");
	        transformer.setOutputProperty(OutputKeys.INDENT, "yes");
	        transformer.setOutputProperty(OutputKeys.ENCODING, "UTF-8");

	        transformer.transform(new DOMSource(doc), new StreamResult(sw));
	        return sw.toString();
	    } catch (Exception ex) {
	        throw new RuntimeException("Error converting to String", ex);
	    }
	}

	/*
	 * Convert the HashMap of testsuites and testcases to xml format write into a file in google cloud.
	 */
	private static void printFlakey(HashMap<String, HashMap<String, Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>>>> fullFlakey, Storage storage, String filePath, String bucketName) throws TransformerException, ParserConfigurationException{

		String xmlPattern = "/^[a-zA-Z_:][a-zA-Z0-9\\.\\-_:]*$/";
		Pattern pattern = Pattern.compile(xmlPattern);


		DocumentBuilderFactory documentFactory = DocumentBuilderFactory.newInstance();
 
        DocumentBuilder documentBuilder = documentFactory.newDocumentBuilder();

        Document document = documentBuilder.newDocument();

        Element root = document.createElement("testsuites");
        document.appendChild(root);
        for (String branchName : fullFlakey.keySet()) {
        	HashMap<String, Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>>> flakey = fullFlakey.get(branchName);
        	Element branch = document.createElement("branch");
        	Attr bran = document.createAttribute("name");
        	bran.setValue(branchName);
        	branch.setAttributeNode(bran);
        	for (String suiteName : flakey.keySet()) {

	        	Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>> result = flakey.get(suiteName);
	        	Pair<Integer, Integer> suiteResult = result.getFirst();
	        	HashMap<String, Pair<Integer, Integer>> caseCollection = result.getSecond();
	        	


	            for (String caseName : caseCollection.keySet()) {
	            	Pair<Integer, Integer> caseResult = caseCollection.get(caseName);
	            	String classAndMethod = caseName.substring(caseName.indexOf("*") + 1);
	            	String className = classAndMethod.substring(0, classAndMethod.indexOf("|"));
	            	String method = classAndMethod.substring(classAndMethod.indexOf("|") + 1);
	            	Element testcase = document.createElement("testcase");

	            	Attr testPathName = document.createAttribute("path");
		            testPathName.setValue(suiteName);
		            testcase.setAttributeNode(testPathName);

	            	Attr testClassName = document.createAttribute("class");
		            testClassName.setValue(className);
		            testcase.setAttributeNode(testClassName);

		            Attr testMethodName = document.createAttribute("method");
		            testMethodName.setValue(method);
		            testcase.setAttributeNode(testMethodName);

	            	Attr caseFailure = document.createAttribute("failures");
		            caseFailure.setValue(Integer.toString(caseResult.getFirst()));
		            testcase.setAttributeNode(caseFailure);

		            Attr caseTotal = document.createAttribute("total");
		            caseTotal.setValue(Integer.toString(caseResult.getSecond()));
		            testcase.setAttributeNode(caseTotal);

		            branch.appendChild(testcase);

	            }

	        	
	        }
	        root.appendChild(branch);
        }

        

        String xmlString = toString(document);

        BlobId blobId = BlobId.of(bucketName, filePath);
	    BlobInfo blobInfo = BlobInfo.newBuilder(blobId).setContentType("text/xml").build();
	    Blob blob = storage.create(blobInfo, xmlString.getBytes(UTF_8));
	}

	/*
	 * Convert the months returned from Date function into parsable strings.
	 */
	private static int convertMonth(String month) {
		if (month.equals("Jan")) {
			return 1;
		} else if (month.equals("Feb")) {
			return 2;
		} else if (month.equals("Mar")) {
			return 3;
		} else if (month.equals("Apr")) {
			return 4;
		} else if (month.equals("May")) {
			return 5;
		} else if (month.equals("Jun")) {
			return 6;
		} else if (month.equals("Jul")) {
			return 7;
		} else if (month.equals("Aug")) {
			return 8;
		} else if (month.equals("Sep")) {
			return 9;
		} else if (month.equals("Oct")) {
			return 10;
		} else if (month.equals("Nov")) {
			return 11;
		} else if (month.equals("Dec")) {
			return 12;
		}
		return 0;
	}

	/*
	 * Compare the date of file to the deadline specified in parameters.
	 */
	private static boolean compareToPast(String date, int days) {
		int day = Integer.parseInt(date.substring(0, date.indexOf(" ")));
		date = date.substring(date.indexOf(" ") + 1);
		int month = convertMonth(date.substring(0, date.indexOf(" ")));
		int year = Integer.parseInt(date.substring(date.indexOf(" ") + 1));

		Calendar cal = Calendar.getInstance();
		cal.add(Calendar.DATE, -days);
		// String for date example: Tue May 14 14:22:48 PDT 2019
		String weekAgo = cal.getTime().toString();
		weekAgo = weekAgo.substring(weekAgo.indexOf(" ") + 1);
		int oldMonth = convertMonth(weekAgo.substring(0, weekAgo.indexOf(" ")));
		weekAgo = weekAgo.substring(weekAgo.indexOf(" ") + 1);
		int oldDay = Integer.parseInt(weekAgo.substring(0, weekAgo.indexOf(" ")));
		int oldYear = Integer.parseInt(weekAgo.substring(weekAgo.lastIndexOf(" ") + 1));

		if (year > oldYear || (year == oldYear && month > oldMonth) || (year == oldYear && month == oldMonth && day >= oldDay)){
			return true;
		}
		return false;

	}

	/*
	 * Read the files according to readInput command.
	 * Select those that qualifies by numDaysPast.
	 * Call other functions to create HashMap of testsuites and cases.
	 * Write result to output file.
	 * Delete the temp folder created with readInput command.
	 */
	public static void testFlakey(int numDaysPast) {
		try {
			String outputFileName = new SimpleDateFormat("dd_MM_yyyy").format(new Date()) + "_" + Integer.toString(numDaysPast) + ".xml";
			// test command path for only with integration test: testCommand.sh
			String contentInput = new String (Files.readAllBytes(Paths.get(pathToReadInput)));
			contentInput = contentInput.replace("$data_folder", dataFolder);
			BufferedWriter writerInput = new BufferedWriter(new FileWriter(pathToReadInput));
    		writerInput.write(contentInput);
    		writerInput.close();
			Process processToRead = Runtime.getRuntime().exec("sh " + pathToReadInput);
			processToRead.waitFor();
			System.out.println("finished running");
			HashMap<String, HashMap<String, Pair<Pair<Integer, Integer>, HashMap<String, Pair<Integer, Integer>>>>> fullFlakey = new HashMap<>();
			
			Storage storage = StorageOptions.getDefaultInstance().getService();
			System.out.println("get storage service");
			
			Page<Blob> blobs =
	     storage.list(
	         bucketName, BlobListOption.currentDirectory(), BlobListOption.prefix(dataFolder + "/"));
	     	System.out.println("get bucket and files");
			for (Blob blob : blobs.iterateAll()) {
				String fileName = blob.getName();
				System.out.println("reading file name: " + fileName);
				System.out.println(fileName);
				
				String fileContent = new String(blob.getContent());
				System.out.println("get content of file");
				String date = fileName.substring(fileName.indexOf("-") + 1);
				date = date.substring(date.indexOf(" ") + 1);
				date = date.substring(0, date.lastIndexOf(" "));
				String branch = fileName.substring(fileName.lastIndexOf("-") + 1);
				branch = branch.substring(0, branch.lastIndexOf(".xml"));
				if (compareToPast(date, numDaysPast)) {
					System.out.println(fileName);
					DocumentBuilder dBuilder = DocumentBuilderFactory.newInstance()
			                             .newDocumentBuilder();
					InputSource is = new InputSource();
					is.setCharacterStream(new StringReader(fileContent));

					Document doc = dBuilder.parse(is);
					identifyFailures(fullFlakey, doc, branch);
				}
			}
			printFlakey(fullFlakey, storage, outputFileName, bucketName);
			System.out.println("write to hash map");
			String content = new String (Files.readAllBytes(Paths.get(pathToDeleteTempCommand)));
			content = content.replace("$data_folder", dataFolder);
			BufferedWriter writer = new BufferedWriter(new FileWriter(pathToDeleteTempCommand));
    		writer.write(content);
    		writer.close();
    		System.out.println("write to storage file");
    		Process processToDelete = Runtime.getRuntime().exec("sh " + pathToDeleteTempCommand);
			processToDelete.waitFor();
			System.out.println("finish deleting temp files");
    		content = new String (Files.readAllBytes(Paths.get(pathToDeleteTempCommand)));
    		content = content.replace(dataFolder, "$data_folder");
    		BufferedWriter newWriter = new BufferedWriter(new FileWriter(pathToDeleteTempCommand));
    		newWriter.write(content);
    		newWriter.close();
    		System.out.println("change the original files");
		} catch (Exception e) {
			System.out.println(e.getMessage());
		}
	}

	public static void main(String[] args) {
		testFlakey(30);
		testFlakey(7);
    }
}




