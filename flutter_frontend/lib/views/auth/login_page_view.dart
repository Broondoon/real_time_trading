import 'package:flutter/material.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _isLoading = false;
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Login'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            TextField(
              controller: _usernameController,
              decoration: const InputDecoration(
                labelText: 'Username',
              )
            ),
            TextField(
              controller: _passwordController,
              decoration: const InputDecoration(
                labelText: 'Password',
              )
            ),
            const SizedBox(
              height: 20,
            ),
            ElevatedButton(
              onPressed: _isLoading ? null : () async {
                print("PRESSED BUTTON TO LOGIN.");
                setState(() => _isLoading = true);
                // final authController = Provider.of<AuthController>(context, listen: false);
                // bool success = await authController.login(
                //   _usernameController.text,
                //   _passwordController.text,
                // );
                bool success = true;
                setState(() => _isLoading = false);
        
                // This is an interesting thing! "Mounted" is whether the current widget
                //    still exists, i.e. is still valid in the build tree.
                // In other words, while we were waiting for the asynch to finish (and isn't that a strange thing to say,
                //    was our widget destroyed during that time?
                if (!context.mounted) return;

                if (success) {
                  context.goNamed(homeRouteName);
                  print("Completed the go.");
                }
                else {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Username and/or password did not match.'),
                    ),
                  );
                }
              },
              child: _isLoading ? const CircularProgressIndicator() : const Text('Login'),
            ),
            ElevatedButton(
              onPressed: () async {
                print("PRESSED BUTTON TO REGISTER.");
                final authController = Provider.of<AuthController>(context, listen: false);
                bool success = await authController.register(
                  _usernameController.text,
                  _passwordController.text,
                );
        
                // This is an interesting thing! "Mounted" is whether the current widget
                //    still exists, i.e. is still valid in the build tree.
                // In other words, while we were waiting for the asynch to finish (and isn't that a strange thing to say,
                //    was our widget destroyed during that time?
                if (!context.mounted) return;

                if (success) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Username and/or password did not match.'),
                    ),
                  );
                }
                else {
                  print("Failed to register - login page");
                }
              },
              child: const Text('Register'),
            )
          ],
        ),
      ),
    );
  }
}

