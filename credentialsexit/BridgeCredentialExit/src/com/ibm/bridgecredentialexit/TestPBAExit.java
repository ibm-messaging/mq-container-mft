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

import static org.junit.Assert.*;

import java.io.File;
import java.net.URL;
import java.util.HashMap;
import java.util.Map;

import org.junit.Test;

import com.ibm.wmqfte.exitroutine.api.CredentialExitResult;
import com.ibm.wmqfte.exitroutine.api.CredentialExitResultCode;
import com.ibm.wmqfte.exitroutine.api.Credentials;
import com.ibm.wmqfte.exitroutine.api.ProtocolServerEndPoint;

/**
 * @author SHASHIKANTHTHAMBRAHA
 *
 */
public class TestPBAExit {

	@Test
	public void testInitializeNoValueProperty() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		
		props.put("protocolBridgeCredentialConfiguration", "");
		assertTrue(exit.initialize(props));
	}

	@Test
	public void testInitializeNoProperty() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();		
		assertTrue(exit.initialize(props));
	}
	
	@Test
	public void testInitializeProperty() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		
		props.put("protocolBridgeCredentialConfiguration", "ProtocolBridgeCredentials.prop");
		assertTrue(exit.initialize(props));
	}

	@Test
	public void testInitializeValidPropertyValue() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
	}
	
	@Test
	public void testValidMapUseridValidSFTPHost() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("elbow", "SFTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "mquserid");
		assertEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
	}

	@Test
	public void testMapUseridNonExistSFTPHost() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("kakapo1", "SFTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "mquserid");
		assertNotEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
	}

	@Test
	public void testMapUseridExistFTPHost() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("mykanos", "FTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "mquserid");
		assertEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
	}

	@Test
	public void testMapUseridExistUnknownV1FTPHost() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.prop");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("mykanos", "FTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "shashikantht");
		assertNotEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
	}

	@Test
	public void testMapUseridExistknownV1FTPHost() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.prop");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("10.17.68.52", "FTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "shashikantht");
		assertEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
		Credentials creds = cer.getCredentials();
		System.out.println("UserId: " + creds.getUserId ());
		assertEquals(creds.getUserId ().get(),"root");
		assertEquals(creds.getPassword ().get(),"Kitt@n0or");
		
		ProtocolServerEndPoint pse1 = new ProtocolServerEndPoint("10.18.68.52", "FTP", "9.122.123.124", 22);
		CredentialExitResult cer1 = exit.mapMQUserId(pse1, "shashikantht");
		assertEquals(cer1.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
		Credentials creds1 = cer1.getCredentials();
		System.out.println("UserId: " + creds.getUserId ());
		assertEquals(creds1.getUserId ().get(),"greekman");
		assertEquals(creds1.getPassword ().get(),"Santorini");
		
	}

	@Test
	public void testMapUseridNoAssocName() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredsNoAssoc.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
	}

	@Test
	public void testMapUseridNoPwd() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredsNoPwd.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("nopassword", "SFTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "shashikantht");
		assertEquals(cer.getResultCode(), CredentialExitResultCode.USER_SUCCESSFULLY_MAPPED);
	}

	@Test
	public void testValidMapUseridValidSFTPHostNonUserMapping() {
		ProtocolBridgeCustomCredentialExit exit = new ProtocolBridgeCustomCredentialExit();
		Map <String, String> props = new HashMap<String, String>();
		URL url = getClass().getResource("ProtocolBridgeCredentials.json");
		File file = new File(url.getPath());
		props.put("protocolBridgeCredentialConfiguration", file.getAbsolutePath());
		assertTrue(exit.initialize(props));
		ProtocolServerEndPoint pse = new ProtocolServerEndPoint("elbow", "SFTP", "9.122.123.124", 22);
		CredentialExitResult cer = exit.mapMQUserId(pse, "nomquserid");
		assertEquals(cer.getResultCode(), CredentialExitResultCode.NO_MAPPING_FOUND);
	}
	
}
