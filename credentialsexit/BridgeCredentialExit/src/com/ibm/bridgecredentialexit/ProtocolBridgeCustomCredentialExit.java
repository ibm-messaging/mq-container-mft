/**
 Copyright IBM Corporation 2020, 2021

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
 */
package com.ibm.wmq.bridgecredentialexit;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.StringTokenizer;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import com.ibm.wmqfte.exitroutine.api.CredentialExitResult;
import com.ibm.wmqfte.exitroutine.api.CredentialExitResultCode;
import com.ibm.wmqfte.exitroutine.api.CredentialHostKey;
import com.ibm.wmqfte.exitroutine.api.CredentialPassword;
import com.ibm.wmqfte.exitroutine.api.CredentialUserId;
import com.ibm.wmqfte.exitroutine.api.Credentials;
import com.ibm.wmqfte.exitroutine.api.ProtocolBridgeCredentialExit2;
import com.ibm.wmqfte.exitroutine.api.ProtocolServerEndPoint;
import com.ibm.wmqfte.exitroutine.api.CredentialPrivateKey;
import org.json.*;

/**
 * A IBM MQ Managed File Transfer Protocol Bridge Custom Credential Exit.
 * This exit reads User Id and Passwords required for a Protocol Bridge Agent
 * to connect to SFTP/FTP server from a file. The User Id and passwords can be
 * in a simple text file or JSON file.
 * 
 */
public class ProtocolBridgeCustomCredentialExit implements ProtocolBridgeCredentialExit2{

	// Internal utility class
	private class CredentialsExt {
		private String transferRequesterId;
		private Credentials credential;

		public CredentialsExt(String requesterId, Credentials cred) {
			this.transferRequesterId = requesterId;
			this.credential = cred;
		}

		public Credentials getCredential() {
			return credential;
		}

		public String getRequesterId() {
			return transferRequesterId;
		}
	}

	// The map that holds mq user ID to serverUserId and serverPassword mappings
	final private Map<String,CredentialsExt> credentialsMap = new HashMap<String, CredentialsExt>();
	final private int ENCODED_PLAIN_TEXT = 0;
	final private int ENCODED_BASE64 = 1;
	private boolean enableDebugLogs = false;

	/* (non-Javadoc)
	 * @see com.ibm.wmqfte.exitroutine.api.ProtocolBridgeCredentialExit#initialize(java.util.Map)
	 */
	@Override
	public boolean initialize(Map<String, String> bridgeProperties) {
		// Get the path of the mq user ID mapping properties file
		String propertiesFilePath = bridgeProperties.get("protocolBridgeCredentialConfiguration");

		if (propertiesFilePath == null || propertiesFilePath.trim().length() == 0) {
			// The properties file path has not been specified. Output an error and return false
			System.err.println("Error initializing custom bridge credentials exit.");
			System.err.println("The location of the mqUserID mapping properties file has not been specified in the protocolBridgeCredentialConfiguration property");
			// Allow agent to come up even if there is an error
			return true;
		}
		
		// Enable debug logs.
		String enableDebugLogStr = System.getenv("ENABLE_PBA_CREDENTIAL_DEBUG_LOG");
		if (enableDebugLogStr != null && enableDebugLogStr.trim().equals("true"))
			enableDebugLogs = true;

		// Trim the whitespaces if any.
		propertiesFilePath = propertiesFilePath.trim();

		// Do some cleanup
		if(!credentialsMap.isEmpty()) {
			credentialsMap.clear();
		}
		
		// Flag to indicate whether the exit has been successfully initialized or not
		boolean initialisationResult = false;
		String fileData = null;

		try {
			File propFile = new File(propertiesFilePath.trim());
			Path path = propFile.toPath();
			BufferedReader reader = Files.newBufferedReader(path);
			StringBuffer sb = new StringBuffer();
			String line = reader.readLine();

			// Read entire file
			while (line != null) {
				sb.append(line);
				line = reader.readLine();
			}
			reader.close();
			fileData = sb.toString();

			// Now parse the data
			if (fileData != null) {
				JSONObject jsonObj = null;
				// First check if the data we got is in JSON format
				try {
					jsonObj = new JSONObject(fileData.trim());
					initialisationResult = processV2Credentials(jsonObj);
				} catch(JSONException jex) {
					if (enableDebugLogs) {
						System.out.println("Failed to load credentials in JSON format from file " +  propertiesFilePath + ". The file does not appear to contain credentials in valid JSON format. Agent will now attempt to decode credentials in V1 (key-value) format");
						System.err.println("Exception details: " + jex);
					}

					// we failed to parse the JSON data, Then this must be old style credential format.
					try {
						initialisationResult  = parseV1FormatCredentials(propertiesFilePath);
					} catch (IOException e) {
						System.err.println("Error loading credentials from file " + propertiesFilePath + " due to exception " + e);
						if (enableDebugLogs) {
							System.err.println(e.getStackTrace());
						}
					}
				} catch(Exception ex) {
					System.err.println("Failed to load credentials due to expection. " + ex);
					if (enableDebugLogs) {
						System.err.println(ex.getStackTrace());
					}
				}
			}
		}
		catch (FileNotFoundException ex) {
			System.err.println("Credentials file " + propertiesFilePath + " not found.");
			if (enableDebugLogs) {
				System.err.println(ex.getStackTrace());
			}
		}
		catch (Exception ex) {
			System.err.println("Failed to initialize custom ProtocolBridgeCustomCredentialExit. The error is: " + ex);
			if (enableDebugLogs) {
				System.err.println(ex.getStackTrace());
			}
		}

		// Allow agent to come up even if there is an error
		return true;
	}

	/**
	 * Read v1 style formatted credentials
	 * @param propertiesFilePath
	 * @throws IOException
	 */
	private boolean parseV1FormatCredentials(String propertiesFilePath) throws IOException {
		boolean initialized = false;
		final Properties mappingProperties = new Properties();
		// Open and load the properties from the properties file
		final File propertiesFile = new File (propertiesFilePath);
		FileInputStream inputStream = null;

		// Create a file input stream to the file
		inputStream = new FileInputStream(propertiesFile);
		// Load the properties from the file
		mappingProperties.load(inputStream);

		// Populate the map of Server Host names to server credentials from the properties
		final Enumeration<?> propertyNames = mappingProperties.propertyNames();
		while ( propertyNames.hasMoreElements()) {
			final Object serverHostName = propertyNames.nextElement();
			if (serverHostName instanceof String ) {
				final String serverHostNameTrim = ((String)serverHostName).trim();
				// Get the value and split into serverUserId and serverPassword 
				final String value = mappingProperties.getProperty(serverHostNameTrim);
				// UID and PWD Credentials will be separated using '!' character.
				final StringTokenizer valueTokenizer = new StringTokenizer(value, "!");

				// Check if we have server type defined as the first token. If present then process accordingly
				// If not defined, then assume FTP as the default.
				if (valueTokenizer.hasMoreTokens()) {
					initialized = processV1FTPCredential(serverHostNameTrim, valueTokenizer);
				}
			}
		}
		return initialized;
	}

	/**
	 * Process multiple V2 formatted Server credentials
	 * @param serverHostName
	 * @param valueTokenizer
	 */
	private boolean processV2Credentials(JSONObject jsonObj) {
		boolean parsed = false;

		JSONArray servers = jsonObj.getJSONArray("servers");
		for(int index = 0; index < servers.length(); index++) {
			JSONObject obj = servers.getJSONObject(index);
			parsed = processV2Credential(obj);
		}

		return parsed;
	}

	/**
	 * Parse single V2 formatted credentials
	 * @param jsonObj
	 * @return
	 */
	private boolean processV2Credential(JSONObject jsonObj) {
		boolean parsed = false;
		String serverType = jsonObj.getString("serverType");
		if(serverType != null) {
			if(serverType.equalsIgnoreCase("SFTP")) {
				parsed = processV2SFTPCredentials(jsonObj);
			} else if(serverType.equalsIgnoreCase("FTPS")) {

			} else if(serverType.equalsIgnoreCase("FTP")) {
				parsed = processV2FTPCredential(jsonObj);
			}
		} else {
			System.out.println("Invalid Server type: " + serverType);
		}

		return parsed;
	}

	/**
	 * Process SFTP Credentials
	 * @param jsonObj
	 * @return
	 */
	private boolean processV2SFTPCredentials(JSONObject jsonObj) {
		String serverHostName = null;
		String serverUserId = null;
		String serverHostKey = null;
		String serverAssocName = null;
		String serverPassword = null;
		String serverPrivateKey = null;
		String requesterUserId = null;
		
		try {
			serverHostName = jsonObj.getString("serverHostName");
		} catch (Exception ex) {
			System.err.println("Credentials does not have the mandatory attribute 'serverHostName' specified. Agent will not be able to communicate with protocol server.");
			return false;
		}

		try {
			serverUserId = jsonObj.getString("serverUserId");
		}catch(Exception ex) {
			System.err.println("Credentials does not have the mandatory attribute 'serverUserId' specified. Agent will not be able to communicate with protocol server.");
			return false;
		}

		try {
			serverPassword = jsonObj.getString("serverPassword");
		}catch(Exception ex) {
			System.err.println("Credentials does not have the mandatory attribute 'serverPassword' specified. Agent will not be able to communicate with protocol server.");
			return false;
		}

		try {
			requesterUserId = jsonObj.getString("transferRequesterId");
		}catch (Exception ex) {
			requesterUserId = "*";
		}

		try {
			serverAssocName = jsonObj.getString("serverAssocName");
		}catch (Exception ex) {
			// Ignore if it's not found.
			serverAssocName = "dummyAssocName";
		}

		try {
			serverPrivateKey = jsonObj.getString("serverPrivateKey");
		}catch (Exception ex) {
			// Not found, ignore
		}

		try {
			serverHostKey = jsonObj.getString("serverHostKey");
		}catch (Exception ex) {
			// Not found, ignore
		}

		String decodedPrivateKey = null;
		List <CredentialPrivateKey> privateKeys = new ArrayList<CredentialPrivateKey>();
		// Private key will be in base64 encoded format. We must decode it now.
		if(serverPrivateKey != null && serverPrivateKey.trim().length() > 0) {
			Base64.Decoder decoder = Base64.getDecoder();  
			// Decode string  
			decodedPrivateKey = new String(decoder.decode(serverPrivateKey));  						
			privateKeys.add(new CredentialPrivateKey(serverAssocName, decodedPrivateKey));
		}

		// Host key also must be in base64 encoding.
		if (serverHostKey != null && serverHostKey.trim().length() > 0) {
			Base64.Decoder decoder = Base64.getDecoder();  
			// Decode string  
			serverHostKey = new String(decoder.decode(serverHostKey));  						
		}

		Credentials credentials = null;
		// If private key is not specified, then use UID/PWD combination.
		if(privateKeys.size() == 0) {
			writeDebugLog("Private key size 0 " + serverHostName + " " + requesterUserId + " " +  serverUserId + " " + serverPassword);
			credentials = new Credentials( new CredentialUserId(serverUserId), new CredentialPassword(serverPassword));
		} else {
			// If server host key is not specified, then use private keys
			if(serverHostKey == null || serverHostKey.isEmpty()) {
				credentials = new Credentials( new CredentialUserId(serverUserId), privateKeys);
			} else {
				credentials = new Credentials( new CredentialUserId(serverUserId), privateKeys, new CredentialHostKey(serverHostKey));
			}

			// Set the password
			credentials.setPassword(new CredentialPassword(serverPassword));
		}
		writeDebugLog("Credential key add to list " + serverHostName + " " + requesterUserId + " " +  serverUserId + " " + serverPassword);
		// Add it to the list
		credentialsMap.put(serverHostName, new CredentialsExt(requesterUserId, credentials));
		return true;
	}

	private boolean processV2FTPCredential(JSONObject jsonObj) {
		boolean valid = true;
		String serverHostName = jsonObj.getString("serverHostName");
		String serverUserId = jsonObj.getString("serverUserId");
		String requesterUserId = jsonObj.getString("transferRequesterId");
		String serverPassword = jsonObj.getString("serverPassword");

		if(serverHostName == null || serverHostName.trim().isEmpty()) {
			valid = false;
		}

		if(valid) {
			if(serverUserId == null || serverUserId.trim().isEmpty()) {
				valid = false;
			}
		}

		if (valid) {
			if(serverPassword == null || serverPassword.trim().isEmpty()) {
				valid = false;
			}
		}

		if(valid) {
			if(requesterUserId == null || requesterUserId.trim().isEmpty()) {
				valid = true;
				requesterUserId = "*";
			}
		}

		if(valid) {
			try {
				// Create a Credential object from the serverUserId and serverPassword
				final Credentials credentials = new Credentials(new CredentialUserId(serverUserId), new CredentialPassword(serverPassword));
				// Insert the credentials into the map
				credentialsMap.put(serverHostName, new CredentialsExt(requesterUserId, credentials));			
			} catch(Exception ex) {
				System.out.println("Failed to create credentials due to exception: " + ex);
				valid = false;
			}
		}

		return valid;
	}

	/**
	 * Process FTP credentials.
	 * @param serverHostName
	 * @param valueTokenizer
	 */
	private boolean processV1FTPCredential(String serverHostName, StringTokenizer valueTokenizer) {
		boolean parsed = false;
		String serverUserId = "";
		String serverPassword = "";
		String encryptionType = "";

		// Server user id
		if (valueTokenizer.hasMoreTokens()) {
			serverUserId = valueTokenizer.nextToken().trim();
		}

		// Type of encoding used.
		if (valueTokenizer.hasMoreTokens()) {
			encryptionType = valueTokenizer.nextToken().trim();
		}

		// Assume the password to be in plain text.
		int encType = ENCODED_PLAIN_TEXT;
		try {
			encType = Integer.parseInt(encryptionType);
		} catch(Exception e) {
			System.err.println(e);
		}

		// Server password.
		if (valueTokenizer.hasMoreTokens()) {
			serverPassword = valueTokenizer.nextToken().trim();
		}

		// If it is base64 encoded. 
		if(encType == ENCODED_BASE64) {
			// Password is base64 encoded. So decode it.
			try {
				byte[] base64decodedBytes = Base64.getDecoder().decode(serverPassword);
				serverPassword = new String(base64decodedBytes, "utf-8");
				// Create a Credential object from the serverUserId and serverPassword
				final Credentials credentials = new Credentials(new CredentialUserId(serverUserId), new CredentialPassword(serverPassword));
				// Insert the credentials into the map
				credentialsMap.put(serverHostName, new CredentialsExt(".*", credentials));
				parsed = true;
			} catch (UnsupportedEncodingException e) {
				System.err.println(e);
			}
		} else if(encType == ENCODED_PLAIN_TEXT) {
			// It's a plain text password
			final Credentials credentials = new Credentials(new CredentialUserId(serverUserId), new CredentialPassword(serverPassword));
			// Insert the credentials into the map
			credentialsMap.put(serverHostName, new CredentialsExt(".*", credentials));
			parsed = true;
		} else {
			// Unknown encoding, don't do anything.
			System.err.println("Unknown encoding " + encType);
		}

		return parsed;
	}


	/* (non-Javadoc)
	 * @see com.ibm.wmqfte.exitroutine.api.ProtocolBridgeCredentialExit#mapMQUserId(java.lang.String)
	 */
	@Override
	public CredentialExitResult mapMQUserId(String mqUserId) {
		CredentialExitResult result = null;
		// Attempt to get the server credentials for the given mq user id
		final CredentialsExt credentials = credentialsMap.get(mqUserId.trim());
		if ( credentials == null) {
			// No entry has been found so return no mapping found with no credentials
			result = new CredentialExitResult(CredentialExitResultCode.NO_MAPPING_FOUND, null);
		}
		else {
			// Some credentials have been found so return success to the user along with the credentials
			result = new CredentialExitResult(CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED, credentials.getCredential());
		}
		return result;
	}

	/* (non-Javadoc)
	 * @see com.ibm.wmqfte.exitroutine.api.ProtocolBridgeCredentialExit#shutdown(java.util.Map)
	 */
	@Override
	public void shutdown(Map<String, String> arg0) {
		// Do some cleanup
		if(!credentialsMap.isEmpty()) {
			credentialsMap.clear();
		}
	}

	/* (non-Javadoc)
	 * @see com.ibm.wmqfte.exitroutine.api.ProtocolBridgeCredentialExit2#mapMQUserId(com.ibm.wmqfte.exitroutine.api.ProtocolServerEndPoint, java.lang.String)
	 */
	@Override
	public CredentialExitResult mapMQUserId(ProtocolServerEndPoint endPointAddress, String mqUserId) {
		CredentialExitResult result = null;
		writeDebugLog("Endpoint Host: " + endPointAddress.getHost() + " Name: " + endPointAddress.getName() + " " + mqUserId);

		// Attempt to get the server credentials for the given mq user id
		final CredentialsExt credentials = credentialsMap.get(endPointAddress.getHost().trim());
		if ( credentials == null) {
			// No entry has been found so return no mapping found with no credentials
			writeLog("Credentials for server" + endPointAddress.getHost()+ "not found ");
			result = new CredentialExitResult(CredentialExitResultCode.NO_MAPPING_FOUND, null);
		}
		else {
			writeDebugLog("Requested Uid " + credentials.getRequesterId());
			// We may need to match the given userId. So do that now.
			if(mqUserId != null && credentials.getRequesterId() != null) {
				// We may have wild cards.
				try {
					String requesterIdCred = credentials.getRequesterId();
					requesterIdCred = requesterIdCred.replaceAll("\\*", "\\\\*");
					Pattern p = Pattern.compile(requesterIdCred);
					Matcher m = p.matcher(mqUserId);
					if(m.matches()) {
						writeDebugLog("Matching Uid found " + requesterIdCred);
						result = new CredentialExitResult(CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED, credentials.getCredential());
					} else {
						// If the requester id is *, then match every id that is supplied
						writeDebugLog("Matching Uid not found " + requesterIdCred);
						if (credentials.getRequesterId().equalsIgnoreCase("*")) {
							result = new CredentialExitResult(CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED, credentials.getCredential());
						} else {
							writeDebugLog("Not generic UID " + requesterIdCred);
							result = new CredentialExitResult(CredentialExitResultCode.NO_MAPPING_FOUND, null);
						}
					}
				} catch (Exception ex) {
					writeDebugLog(ex.toString());
					result = new CredentialExitResult(CredentialExitResultCode.NO_MAPPING_FOUND, null);
				}
			} else {
				writeLog("Requester id and mq userid not matching ");
				// Some credentials have been found so return success to the user along with the credentials
				result = new CredentialExitResult(CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED, credentials.getCredential());
			}
		}
		return result;
	}

	private void writeLog(String log) {
		System.out.println(log);
	}

	// Write a debug log to agent's output0.log
	private void writeDebugLog(String log) {
		if (enableDebugLogs)
			System.out.println(log);
	}
}
